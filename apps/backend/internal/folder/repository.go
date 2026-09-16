package folder

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

func (r *Repository) Create(ctx context.Context, params db.CreateFolderParams) (db.Folder, error) {
	return r.queries.CreateFolder(ctx, params)
}

func (r *Repository) GetByID(ctx context.Context, id string) (db.Folder, error) {
	return r.queries.GetFolderByID(ctx, id)
}

func (r *Repository) GetByIDAndOwner(ctx context.Context, id, ownerID string) (db.Folder, error) {
	return r.queries.GetFolderByIDAndOwner(ctx, db.GetFolderByIDAndOwnerParams{
		ID:      id,
		OwnerID: ownerID,
	})
}

func (r *Repository) ListRoot(ctx context.Context, ownerID string) ([]db.Folder, error) {
	return r.queries.ListRootFolders(ctx, ownerID)
}

func (r *Repository) ListByParent(ctx context.Context, ownerID string, parentFolderID sql.NullString) ([]db.Folder, error) {
	return r.queries.ListFoldersByParent(ctx, db.ListFoldersByParentParams{
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
	})
}

func (r *Repository) GetExistingNamesInRoot(ctx context.Context, ownerID string) ([]string, error) {
	return r.queries.GetExistingFolderNamesInRoot(ctx, ownerID)
}

func (r *Repository) GetExistingNamesInParent(ctx context.Context, ownerID string, parentFolderID sql.NullString) ([]string, error) {
	return r.queries.GetExistingFolderNamesInParent(ctx, db.GetExistingFolderNamesInParentParams{
		OwnerID:        ownerID,
		ParentFolderID: parentFolderID,
	})
}

func (r *Repository) Rename(ctx context.Context, params db.RenameFolderParams) (db.Folder, error) {
	return r.queries.RenameFolder(ctx, params)
}

func (r *Repository) Move(ctx context.Context, params db.MoveFolderParams) (db.Folder, error) {
	return r.queries.MoveFolder(ctx, params)
}

func (r *Repository) SoftDelete(ctx context.Context, params db.SoftDeleteFolderParams) (db.Folder, error) {
	return r.queries.SoftDeleteFolder(ctx, params)
}

func (r *Repository) SoftDeleteByID(ctx context.Context, params db.SoftDeleteFolderByIDParams) error {
	return r.queries.SoftDeleteFolderByID(ctx, params)
}

func (r *Repository) GetDescendantIDs(ctx context.Context, id, ownerID string) ([]string, error) {
	return r.queries.GetFolderDescendantIDs(ctx, db.GetFolderDescendantIDsParams{
		ID:      id,
		OwnerID: ownerID,
	})
}

func (r *Repository) Restore(ctx context.Context, params db.RestoreFolderParams) (db.Folder, error) {
	return r.queries.RestoreFolder(ctx, params)
}

func (r *Repository) PermanentDelete(ctx context.Context, params db.PermanentDeleteFolderParams) error {
	return r.queries.PermanentDeleteFolder(ctx, params)
}

func (r *Repository) SetStarred(ctx context.Context, params db.SetFolderStarredParams) (db.Folder, error) {
	return r.queries.SetFolderStarred(ctx, params)
}

func (r *Repository) ListStarred(ctx context.Context, ownerID string) ([]db.Folder, error) {
	return r.queries.ListStarredFolders(ctx, ownerID)
}

func (r *Repository) ListTrash(ctx context.Context, ownerID string) ([]db.Folder, error) {
	return r.queries.ListTrashFolders(ctx, ownerID)
}

func (r *Repository) GetStorageKeysByFolderID(ctx context.Context, parentFolderID sql.NullString, ownerID string) ([]string, error) {
	return r.queries.GetStorageKeysByFolderID(ctx, db.GetStorageKeysByFolderIDParams{
		ParentFolderID: parentFolderID,
		OwnerID:        ownerID,
	})
}
