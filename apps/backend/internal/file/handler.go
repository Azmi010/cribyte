package file

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azmi010/my-drive/apps/backend/internal/auth"
	"github.com/Azmi010/my-drive/apps/backend/internal/storage"
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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
