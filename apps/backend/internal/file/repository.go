package file

import (
	"context"
	"database/sql"

	"github.com/Azmi010/cribyte/apps/backend/internal/db"
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Create(ctx context.Context, params db.CreateFileParams) (db.File, error) {
	return r.queries.CreateFile(ctx, params)
}

func (r *Repository) GetByID(ctx context.Context, id string) (db.File, error) {
	return r.queries.GetFileByID(ctx, id)
}

func (r *Repository) GetByIDAndOwner(ctx context.Context, id, ownerID string) (db.File, error) {
	return r.queries.GetFileByIDAndOwner(ctx, db.GetFileByIDAndOwnerParams{
		ID:      id,
		OwnerID: ownerID,
	})
}

func (r *Repository) ListRoot(ctx context.Context, ownerID string) ([]db.File, error) {
	return r.queries.ListRootFiles(ctx, ownerID)
}

func (r *Repository) ListByParent(ctx context.Context, ownerID string, parentFolderID sql.NullString) ([]db.File, error) {
	return r.queries.ListFilesByParent(ctx, db.ListFilesByParentParams{
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
	})
}

func (r *Repository) GetExistingNamesInRoot(ctx context.Context, ownerID string) ([]string, error) {
	return r.queries.GetExistingFileNamesInRoot(ctx, ownerID)
}

func (r *Repository) GetExistingNamesInParent(ctx context.Context, ownerID string, parentFolderID sql.NullString) ([]string, error) {
	return r.queries.GetExistingFileNamesInParent(ctx, db.GetExistingFileNamesInParentParams{
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
	})
}

func (r *Repository) Rename(ctx context.Context, params db.RenameFileParams) (db.File, error) {
	return r.queries.RenameFile(ctx, params)
}

func (r *Repository) Move(ctx context.Context, params db.MoveFileParams) (db.File, error) {
	return r.queries.MoveFile(ctx, params)
}

func (r *Repository) SoftDelete(ctx context.Context, params db.SoftDeleteFileParams) (db.File, error) {
	return r.queries.SoftDeleteFile(ctx, params)
}

func (r *Repository) Restore(ctx context.Context, params db.RestoreFileParams) (db.File, error) {
	return r.queries.RestoreFile(ctx, params)
}

func (r *Repository) PermanentDelete(ctx context.Context, params db.PermanentDeleteFileParams) error {
	return r.queries.PermanentDeleteFile(ctx, params)
}

func (r *Repository) SetStarred(ctx context.Context, params db.SetFileStarredParams) (db.File, error) {
	return r.queries.SetFileStarred(ctx, params)
}

func (r *Repository) SearchByName(ctx context.Context, params db.SearchFilesByNameParams) ([]db.File, error) {
	return r.queries.SearchFilesByName(ctx, params)
}

func (r *Repository) ListStarred(ctx context.Context, ownerID string) ([]db.File, error) {
	return r.queries.ListStarredFiles(ctx, ownerID)
}

func (r *Repository) ListTrash(ctx context.Context, ownerID string) ([]db.File, error) {
	return r.queries.ListTrashFiles(ctx, ownerID)
}

func (r *Repository) GetStorageKeysByFolderID(ctx context.Context, parentFolderID sql.NullString, ownerID string) ([]string, error) {
	return r.queries.GetStorageKeysByFolderID(ctx, db.GetStorageKeysByFolderIDParams{
		ParentFolderID: parentFolderID,
		OwnerID:        ownerID,
	})
}
