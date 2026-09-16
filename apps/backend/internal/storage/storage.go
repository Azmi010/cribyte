package storage

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Stat(ctx context.Context, key string) (int64, error)
	Path(key string) (string, error)
}
