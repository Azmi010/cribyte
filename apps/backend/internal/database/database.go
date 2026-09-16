package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Azmi010/cribyte/apps/backend/internal/config"
	"github.com/Azmi010/cribyte/apps/backend/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Connect establishes a connection to SQLite or PostgreSQL based on the config.
func Connect(cfg *config.Config) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)

	switch cfg.DBDriver {
	case "postgres":
		db, err = sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres connection: %w", err)
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

	case "sqlite", "":
		// Ensure the directory exists
		dir := filepath.Dir(cfg.DatabaseURL)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create sqlite database directory: %w", err)
			}
		}

		db, err = sql.Open("sqlite", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}

		// Configure SQLite pragmas for performance and data integrity
		pragmas := []string{
			"PRAGMA foreign_keys = ON;",
			"PRAGMA journal_mode = WAL;",
			"PRAGMA busy_timeout = 5000;",
			"PRAGMA synchronous = NORMAL;",
		}
		for _, pragma := range pragmas {
			if _, err := db.Exec(pragma); err != nil {
				return nil, fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
			}
		}

		// SQLite works best with 1 open writer or managed connection pool
		db.SetMaxOpenConns(1)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DBDriver)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	slog.Info("connected to database", "driver", cfg.DBDriver)
	return db, nil
}

// Migrate applies all embedded migrations to the database.
func Migrate(db *sql.DB, driver string) error {
	dialect := "sqlite3"
	if driver == "postgres" {
		dialect = "postgres"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	goose.SetBaseFS(migrations.FS)

	slog.Info("running database migrations...")
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("database migrations applied successfully")
	return nil
}
