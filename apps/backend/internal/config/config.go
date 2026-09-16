package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	// Database
	DBDriver    string // "sqlite" or "postgres"
	DatabaseURL string // path to sqlite file or postgres connection string

	// Storage
	StorageDriver    string // "local" or "s3"
	StorageLocalPath string
	StorageS3        S3Config

	// Auth / Session
	SessionSecret      string
	SessionExpiryHours int
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "4000"),
		Env:                getEnv("ENV", "development"),
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
		DatabaseURL:        getEnv("DATABASE_URL", "./data/cribyte.db"),
		StorageDriver:      getEnv("STORAGE_DRIVER", "local"),
		StorageLocalPath:   getEnv("STORAGE_LOCAL_PATH", "./data/uploads"),
		SessionSecret:      getEnv("SESSION_SECRET", "default-dev-secret-key-change-me"),
		SessionExpiryHours: getEnvInt("SESSION_EXPIRY_HOURS", 168),
		StorageS3: S3Config{
			Endpoint:  getEnv("STORAGE_S3_ENDPOINT", "localhost:9000"),
			Bucket:    getEnv("STORAGE_S3_BUCKET", "cribyte"),
			AccessKey: getEnv("STORAGE_S3_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("STORAGE_S3_SECRET_KEY", "minioadmin"),
			UseSSL:    getEnvBool("STORAGE_S3_USE_SSL", false),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
