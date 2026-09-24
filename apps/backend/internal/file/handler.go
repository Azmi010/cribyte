package file

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azmi010/cribyte/apps/backend/internal/auth"
	"github.com/Azmi010/cribyte/apps/backend/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc     *Service
	storage storage.Storage
}

func NewHandler(svc *Service, storage storage.Storage) *Handler {
	return &Handler{svc: svc, storage: storage}
}

type renameRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type moveRequest struct {
	ParentFolderID *string `json:"parent_folder_id"`
}

type starredRequest struct {
	Starred bool `json:"starred"`
}

// Upload godoc
// @Summary Upload a file
// @Description Upload a file via multipart form. Optionally specify parent_folder_id and custom name.
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security SessionAuth
// @Param file formData file true "File to upload"
// @Param parent_folder_id formData string false "Parent folder ID"
// @Param name formData string false "Custom file name"
// @Success 201 {object} file.UploadResult
// @Failure 400 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Router /api/files/upload [post]
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize+1<<20)

	if err := r.ParseMultipartForm(MaxUploadSize + 1<<20); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			writeError(w, http.StatusRequestEntityTooLarge, ErrTooLarge.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	fileHeader, handler, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer fileHeader.Close()

	var parentFolderID *string
	if pid := r.FormValue("parent_folder_id"); pid != "" {
		parentFolderID = &pid
	}

	name := handler.Filename
	if customName := r.FormValue("name"); customName != "" {
		name = customName
	}

	result, err := h.svc.Upload(r.Context(), user.ID, name, parentFolderID, fileHeader, handler.Size)
	if err != nil {
		switch {
		case errors.Is(err, ErrTooLarge):
			writeError(w, http.StatusRequestEntityTooLarge, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetByID godoc
// @Summary Get file metadata
// @Description Get metadata for a file by its ID
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {object} file.FileResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	result, err := h.svc.GetByID(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListContents godoc
// @Summary List files
// @Description List files in a folder or search by name
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param folder_id query string false "Parent folder ID"
// @Param search query string false "Search query"
// @Success 200 {array} file.FileResult
// @Failure 401 {object} map[string]string
// @Router /api/files [get]
func (h *Handler) ListContents(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var parentFolderID *string
	if pid := r.URL.Query().Get("folder_id"); pid != "" {
		parentFolderID = &pid
	}

	if q := r.URL.Query().Get("search"); q != "" {
		results, err := h.svc.Search(r.Context(), user.ID, q)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, results)
		return
	}

	results, err := h.svc.ListContents(r.Context(), user.ID, parentFolderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// Download godoc
// @Summary Download a file
// @Description Download file content as attachment
// @Tags files
// @Produce octet-stream
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {file} binary
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/download [get]
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	reader, file, err := h.svc.Download(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))

	io.Copy(w, reader)
}

// Preview godoc
// @Summary Preview a file
// @Description Get preview info (kind, mime_type, content for text files, signed URL for others)
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {object} file.PreviewResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/preview [get]
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	result, err := h.svc.Preview(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ServeContent godoc
// @Summary Serve file content inline
// @Description Serve file content directly (supports Range requests for video/audio)
// @Tags files
// @Produce octet-stream
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {file} binary
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/serve [get]
func (h *Handler) ServeContent(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	storageKey, err := h.svc.GetStorageKey(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	filePath, err := h.storage.Path(storageKey)
	if err != nil {
		signedURL, err2 := h.storage.URL(r.Context(), storageKey, 1*time.Hour)
		if err2 != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		http.Redirect(w, r, signedURL, http.StatusTemporaryRedirect)
		return
	}

	http.ServeFile(w, r, filePath)
}

// Thumbnail godoc
// @Summary Get file thumbnail
// @Description Get a small JPEG thumbnail for image, video, or PDF files. Generated and cached on first request.
// @Tags files
// @Produce jpeg
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {file} binary
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/thumbnail [get]
func (h *Handler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	data, err := h.svc.Thumbnail(r.Context(), id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrNoThumbnail):
			writeError(w, http.StatusNotFound, "no thumbnail available")
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = w.Write(data)
}

// Rename godoc
// @Summary Rename a file
// @Description Rename a file. Auto-resolves name conflicts.
// @Tags files
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Param request body renameRequest true "New name"
// @Success 200 {object} file.FileResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id} [patch]
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	result, err := h.svc.Rename(r.Context(), id, user.ID, RenameInput{
		Name: req.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrNameConflict):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Move godoc
// @Summary Move a file
// @Description Move a file to a different folder
// @Tags files
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Param request body moveRequest true "Target folder"
// @Success 200 {object} file.FileResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/move [post]
func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	var req moveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Move(r.Context(), id, user.ID, MoveInput{
		ParentFolderID: req.ParentFolderID,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Trash godoc
// @Summary Trash a file
// @Description Soft-delete a file (move to trash)
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {object} file.FileResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/trash [delete]
func (h *Handler) Trash(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	result, err := h.svc.Trash(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Restore godoc
// @Summary Restore a file from trash
// @Description Restore a trashed file to its original location
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {object} file.FileResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/restore [post]
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	result, err := h.svc.Restore(r.Context(), id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// PermanentDelete godoc
// @Summary Permanently delete a file
// @Description Delete a file permanently from trash. File must be in trash first.
// @Tags files
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/permanent-delete [post]
func (h *Handler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.svc.PermanentDelete(r.Context(), id, user.ID); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "file permanently deleted"})
}

// ToggleStarred godoc
// @Summary Star or unstar a file
// @Description Toggle the starred/favorite status of a file
// @Tags files
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "File ID"
// @Param request body starredRequest true "Starred status"
// @Success 200 {object} file.FileResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id}/star [post]
func (h *Handler) ToggleStarred(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")

	var req starredRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.ToggleStarred(r.Context(), id, user.ID, req.Starred)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListStarred godoc
// @Summary List starred files
// @Description List all starred files for the current user
// @Tags files
// @Produce json
// @Security SessionAuth
// @Success 200 {array} file.FileResult
// @Failure 401 {object} map[string]string
// @Router /api/files/starred [get]
func (h *Handler) ListStarred(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	results, err := h.svc.ListStarred(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// ListTrash godoc
// @Summary List trashed files
// @Description List all files in trash for the current user
// @Tags files
// @Produce json
// @Security SessionAuth
// @Success 200 {array} file.FileResult
// @Failure 401 {object} map[string]string
// @Router /api/files/trash [get]
func (h *Handler) ListTrash(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	results, err := h.svc.ListTrash(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
