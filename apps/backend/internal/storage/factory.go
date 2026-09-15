package storage

import (
	"fmt"

	"github.com/Azmi010/my-drive/apps/backend/internal/config"
)

func New(cfg *config.Config) (Storage, error) {
	switch cfg.StorageDriver {
	case "local":
		return NewLocalStorage(cfg.StorageLocalPath, cfg.SessionSecret, "/api/files/serve")
	case "s3":
		return NewS3Storage(
			cfg.StorageS3.Endpoint,
			cfg.StorageS3.AccessKey,
			cfg.StorageS3.SecretKey,
			cfg.StorageS3.Bucket,
			cfg.StorageS3.UseSSL,
		)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.StorageDriver)
	}
}
