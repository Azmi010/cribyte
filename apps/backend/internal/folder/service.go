package folder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Azmi010/cribyte/apps/backend/internal/db"
	"github.com/Azmi010/cribyte/apps/backend/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("folder not found")
	ErrNameConflict  = errors.New("folder with this name already exists")
	ErrParentNotFound = errors.New("parent folder not found")
	ErrMoveToChild   = errors.New("cannot move folder into its own descendant")
)

type Service struct {
	repo    *Repository
	storage storage.Storage
}

func NewService(repo *Repository, storage storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

type CreateInput struct {
	Name           string
	ParentFolderID *string // nil = root
}

type RenameInput struct {
	Name string
}

type MoveInput struct {
	ParentFolderID *string // nil = root
}

type FolderResult struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Starred        bool      `json:"starred"`
	ParentFolderID *string   `json:"parent_folder_id"`
	OwnerID        string    `json:"owner_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toResult(f db.Folder) FolderResult {
	var parentID *string
	if f.ParentFolderID.Valid {
		parentID = &f.ParentFolderID.String
	}
	return FolderResult{
		ID:             f.ID,
		Name:           f.Name,
		Starred:        f.Starred,
		ParentFolderID: parentID,
		OwnerID:        f.OwnerID,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

func (s *Service) Create(ctx context.Context, ownerID string, input CreateInput) (FolderResult, error) {
	var parentID sql.NullString
	if input.ParentFolderID != nil {
		parentID = sql.NullString{String: *input.ParentFolderID, Valid: true}

		_, err := s.repo.GetByIDAndOwner(ctx, *input.ParentFolderID, ownerID)
		if err != nil {
			return FolderResult{}, ErrParentNotFound
		}
	}

	name := s.resolveUniqueName(ctx, ownerID, input.Name, parentID)

	now := time.Now()
	folder, err := s.repo.Create(ctx, db.CreateFolderParams{
		ID:             uuid.New().String(),
		Name:           name,
		ParentFolderID: parentID,
		OwnerID:        ownerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		slog.Error("failed to create folder", "error", err)
		return FolderResult{}, err
	}

	slog.Info("folder created", "id", folder.ID, "name", folder.Name)
	return toResult(folder), nil
}

func (s *Service) GetByID(ctx context.Context, id, ownerID string) (FolderResult, error) {
	folder, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}
	return toResult(folder), nil
}

func (s *Service) ListContents(ctx context.Context, ownerID string, parentFolderID *string) ([]FolderResult, error) {
	var parentID sql.NullString
	if parentFolderID != nil {
		parentID = sql.NullString{String: *parentFolderID, Valid: true}

		_, err := s.repo.GetByIDAndOwner(ctx, *parentFolderID, ownerID)
		if err != nil {
			return nil, ErrNotFound
		}
	}

	var folders []db.Folder
	var err error
	if parentFolderID == nil {
		folders, err = s.repo.ListRoot(ctx, ownerID)
	} else {
		folders, err = s.repo.ListByParent(ctx, ownerID, parentID)
	}
	if err != nil {
		slog.Error("failed to list folders", "error", err)
		return nil, err
	}

	results := make([]FolderResult, len(folders))
	for i, f := range folders {
		results[i] = toResult(f)
	}
	return results, nil
}

func (s *Service) Rename(ctx context.Context, id, ownerID string, input RenameInput) (FolderResult, error) {
	folder, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}

	name := s.resolveUniqueNameForRename(ctx, ownerID, input.Name, folder.ParentFolderID, id)

	updated, err := s.repo.Rename(ctx, db.RenameFolderParams{
		Name:      name,
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to rename folder", "error", err)
		return FolderResult{}, err
	}

	return toResult(updated), nil
}

func (s *Service) Move(ctx context.Context, id, ownerID string, input MoveInput) (FolderResult, error) {
	_, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}

	var targetParentID sql.NullString
	if input.ParentFolderID != nil {
		targetParentID = sql.NullString{String: *input.ParentFolderID, Valid: true}

		_, err := s.repo.GetByIDAndOwner(ctx, *input.ParentFolderID, ownerID)
		if err != nil {
			return FolderResult{}, ErrParentNotFound
		}

		if *input.ParentFolderID == id {
			return FolderResult{}, ErrMoveToChild
		}

		descendants, err := s.repo.GetDescendantIDs(ctx, id, ownerID)
		if err != nil {
			slog.Error("failed to get descendants", "error", err)
			return FolderResult{}, err
		}
		for _, descID := range descendants {
			if descID == *input.ParentFolderID {
				return FolderResult{}, ErrMoveToChild
			}
		}
	}

	updated, err := s.repo.Move(ctx, db.MoveFolderParams{
		ParentFolderID: targetParentID,
		UpdatedAt:      time.Now(),
		ID:             id,
		OwnerID:        ownerID,
	})
	if err != nil {
		slog.Error("failed to move folder", "error", err)
		return FolderResult{}, err
	}

	return toResult(updated), nil
}

func (s *Service) Trash(ctx context.Context, id, ownerID string) (FolderResult, error) {
	_, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}

	now := time.Now()
	nullTime := sql.NullTime{Time: now, Valid: true}

	descendants, err := s.repo.GetDescendantIDs(ctx, id, ownerID)
	if err != nil {
		slog.Error("failed to get descendants for trash", "error", err)
		return FolderResult{}, err
	}

	for _, descID := range descendants {
		if err := s.repo.SoftDeleteByID(ctx, db.SoftDeleteFolderByIDParams{
			DeletedAt: nullTime,
			UpdatedAt: now,
			ID:        descID,
			OwnerID:   ownerID,
		}); err != nil {
			slog.Error("failed to trash descendant folder", "id", descID, "error", err)
		}

		var descParentID sql.NullString
		if descID != id {
			descParentID = sql.NullString{String: descID, Valid: true}
		} else {
			descParentID = sql.NullString{String: id, Valid: true}
		}

		storageKeys, err := s.repo.GetStorageKeysByFolderID(ctx, descParentID, ownerID)
		if err != nil {
			slog.Error("failed to get storage keys for trash", "folder_id", descID, "error", err)
			continue
		}
		for _, key := range storageKeys {
			if err := s.storage.Delete(ctx, key); err != nil {
				slog.Error("failed to delete storage key", "key", key, "error", err)
			}
		}
	}

	folderStorageKeys, err := s.repo.GetStorageKeysByFolderID(ctx, sql.NullString{String: id, Valid: true}, ownerID)
	if err == nil {
		for _, key := range folderStorageKeys {
			if err := s.storage.Delete(ctx, key); err != nil {
				slog.Error("failed to delete storage key", "key", key, "error", err)
			}
		}
	}

	result, err := s.repo.SoftDelete(ctx, db.SoftDeleteFolderParams{
		DeletedAt: nullTime,
		UpdatedAt: now,
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to trash folder", "error", err)
		return FolderResult{}, err
	}

	slog.Info("folder trashed", "id", id)
	return toResult(result), nil
}

func (s *Service) Restore(ctx context.Context, id, ownerID string) (FolderResult, error) {
	folder, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}

	if !folder.Deleted {
		return FolderResult{}, errors.New("folder is not in trash")
	}

	result, err := s.repo.Restore(ctx, db.RestoreFolderParams{
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to restore folder", "error", err)
		return FolderResult{}, err
	}

	slog.Info("folder restored", "id", id)
	return toResult(result), nil
}

func (s *Service) PermanentDelete(ctx context.Context, id, ownerID string) error {
	folder, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return ErrNotFound
	}

	if !folder.Deleted {
		return errors.New("folder must be in trash before permanent delete")
	}

	descendants, err := s.repo.GetDescendantIDs(ctx, id, ownerID)
	if err != nil {
		slog.Error("failed to get descendants for permanent delete", "error", err)
		return err
	}

	for _, descID := range descendants {
		storageKeys, err := s.repo.GetStorageKeysByFolderID(ctx, sql.NullString{String: descID, Valid: true}, ownerID)
		if err != nil {
			slog.Error("failed to get storage keys for permanent delete", "folder_id", descID, "error", err)
			continue
		}
		for _, key := range storageKeys {
			if err := s.storage.Delete(ctx, key); err != nil {
				slog.Error("failed to delete storage key", "key", key, "error", err)
			}
		}
	}

	folderStorageKeys, err := s.repo.GetStorageKeysByFolderID(ctx, sql.NullString{String: id, Valid: true}, ownerID)
	if err == nil {
		for _, key := range folderStorageKeys {
			if err := s.storage.Delete(ctx, key); err != nil {
				slog.Error("failed to delete storage key", "key", key, "error", err)
			}
		}
	}

	if err := s.repo.PermanentDelete(ctx, db.PermanentDeleteFolderParams{
		ID:      id,
		OwnerID: ownerID,
	}); err != nil {
		slog.Error("failed to permanent delete folder", "error", err)
		return err
	}

	slog.Info("folder permanently deleted", "id", id)
	return nil
}

func (s *Service) ToggleStarred(ctx context.Context, id, ownerID string, starred bool) (FolderResult, error) {
	folder, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FolderResult{}, ErrNotFound
	}

	result, err := s.repo.SetStarred(ctx, db.SetFolderStarredParams{
		Starred:   starred,
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to toggle starred", "error", err)
		return FolderResult{}, err
	}

	_ = folder
	return toResult(result), nil
}

func (s *Service) ListStarred(ctx context.Context, ownerID string) ([]FolderResult, error) {
	folders, err := s.repo.ListStarred(ctx, ownerID)
	if err != nil {
		slog.Error("failed to list starred folders", "error", err)
		return nil, err
	}

	results := make([]FolderResult, len(folders))
	for i, f := range folders {
		results[i] = toResult(f)
	}
	return results, nil
}

func (s *Service) ListTrash(ctx context.Context, ownerID string) ([]FolderResult, error) {
	folders, err := s.repo.ListTrash(ctx, ownerID)
	if err != nil {
		slog.Error("failed to list trash folders", "error", err)
		return nil, err
	}

	results := make([]FolderResult, len(folders))
	for i, f := range folders {
		results[i] = toResult(f)
	}
	return results, nil
}

func (s *Service) resolveUniqueName(ctx context.Context, ownerID, name string, parentFolderID sql.NullString) string {
	var existing []string
	var err error

	if parentFolderID.Valid {
		existing, err = s.repo.GetExistingNamesInParent(ctx, ownerID, parentFolderID)
	} else {
		existing, err = s.repo.GetExistingNamesInRoot(ctx, ownerID)
	}
	if err != nil {
		return name
	}

	return findAvailableName(name, existing)
}

func (s *Service) resolveUniqueNameForRename(ctx context.Context, ownerID, name string, parentFolderID sql.NullString, excludeID string) string {
	var existing []string
	var err error

	if parentFolderID.Valid {
		existing, err = s.repo.GetExistingNamesInParent(ctx, ownerID, parentFolderID)
	} else {
		existing, err = s.repo.GetExistingNamesInRoot(ctx, ownerID)
	}
	if err != nil {
		return name
	}

	filtered := make([]string, 0, len(existing))
	for _, n := range existing {
		if n != name {
			filtered = append(filtered, n)
		}
	}

	return findAvailableName(name, filtered)
}

func findAvailableName(name string, existing []string) string {
	existingSet := make(map[string]bool, len(existing))
	for _, n := range existing {
		existingSet[n] = true
	}

	if !existingSet[name] {
		return name
	}

	base := name
	ext := ""
	if idx := strings.LastIndex(name, "."); idx > 0 {
		base = name[:idx]
		ext = name[idx:]
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if !existingSet[candidate] {
			return candidate
		}
	}
}
