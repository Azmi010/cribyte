package folder

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Azmi010/my-drive/apps/backend/internal/auth"
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

	id := chi.URLParam(r, "id")

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

	writeJSON(w, http.StatusOK, map[string]string{"message": "folder permanently deleted"})
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
