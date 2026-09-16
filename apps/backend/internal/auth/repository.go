package auth

import (
	"context"

	"github.com/Azmi010/cribyte/apps/backend/internal/db"
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Create(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(ctx, params)
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *Repository) GetByID(ctx context.Context, id string) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}
