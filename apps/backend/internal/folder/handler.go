package folder

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Azmi010/cribyte/apps/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=255"`
	ParentFolderID *string `json:"parent_folder_id"`
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

// Create godoc
// @Summary Create a folder
// @Description Create a new folder. Auto-resolves name conflicts.
// @Tags folders
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body createRequest true "Folder name and optional parent"
// @Success 201 {object} folder.FolderResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	result, err := h.svc.Create(r.Context(), user.ID, CreateInput{
		Name:           req.Name,
		ParentFolderID: req.ParentFolderID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrParentNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrNameConflict):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetByID godoc
// @Summary Get folder metadata
// @Description Get metadata for a folder by its ID
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Success 200 {object} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id} [get]
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
// @Summary List folder contents
// @Description List subfolders inside a folder
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Success 200 {array} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/contents [get]
func (h *Handler) ListContents(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("folder_id")
	}

	var parentID *string
	if id != "" {
		parentID = &id
	}

	results, err := h.svc.ListContents(r.Context(), user.ID, parentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// Rename godoc
// @Summary Rename a folder
// @Description Rename a folder. Auto-resolves name conflicts.
// @Tags folders
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Param request body renameRequest true "New name"
// @Success 200 {object} folder.FolderResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id} [patch]
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
// @Summary Move a folder
// @Description Move a folder to a different parent
// @Tags folders
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Param request body moveRequest true "Target parent folder"
// @Success 200 {object} folder.FolderResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/move [post]
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
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrParentNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrMoveToChild):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Trash godoc
// @Summary Trash a folder
// @Description Soft-delete a folder and all its descendants
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Success 200 {object} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/trash [delete]
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
// @Summary Restore a folder from trash
// @Description Restore a trashed folder to its original location
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Success 200 {object} folder.FolderResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/restore [post]
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
// @Summary Permanently delete a folder
// @Description Delete a folder permanently from trash. Folder must be in trash first.
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/permanent-delete [post]
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "folder permanently deleted"})
}

// ToggleStarred godoc
// @Summary Star or unstar a folder
// @Description Toggle the starred/favorite status of a folder
// @Tags folders
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Folder ID"
// @Param request body starredRequest true "Starred status"
// @Success 200 {object} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/folders/{id}/star [post]
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
// @Summary List starred folders
// @Description List all starred folders for the current user
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Success 200 {array} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Router /api/folders/starred [get]
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
// @Summary List trashed folders
// @Description List all folders in trash for the current user
// @Tags folders
// @Produce json
// @Security SessionAuth
// @Success 200 {array} folder.FolderResult
// @Failure 401 {object} map[string]string
// @Router /api/folders/trash [get]
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
