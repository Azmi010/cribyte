package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azmi010/cribyte/apps/backend/internal/auth"
	"github.com/Azmi010/cribyte/apps/backend/internal/config"
	"github.com/Azmi010/cribyte/apps/backend/internal/database"
	"github.com/Azmi010/cribyte/apps/backend/internal/db"
	"github.com/Azmi010/cribyte/apps/backend/internal/file"
	"github.com/Azmi010/cribyte/apps/backend/internal/folder"
	"github.com/Azmi010/cribyte/apps/backend/internal/ratelimit"
	"github.com/Azmi010/cribyte/apps/backend/internal/session"
	"github.com/Azmi010/cribyte/apps/backend/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/Azmi010/cribyte/apps/backend/docs"
)

// @title CriByte API
// @version 1.0
// @description CriByte — self-hosted Google Drive clone
// @host localhost:4000
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session_id
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	var logHandler slog.Handler
	if cfg.Env == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(logHandler))

	slog.Info("starting CriByte backend", "env", cfg.Env, "port", cfg.Port)

	dbConn, err := database.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	if err := database.Migrate(dbConn, cfg.DBDriver); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	queries := db.New(dbConn)
	sessions := session.NewMemoryStore()

	authRepo := auth.NewRepository(queries)
	authSvc := auth.NewService(authRepo, sessions, cfg)
	authHandler := auth.NewHandler(authSvc)
	authMiddleware := auth.Middleware(sessions, authSvc)

	storageDriver, err := storage.New(cfg)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	folderRepo := folder.NewRepository(queries)
	folderSvc := folder.NewService(folderRepo, storageDriver)
	folderHandler := folder.NewHandler(folderSvc)

	fileRepo := file.NewRepository(queries)
	fileSvc := file.NewService(fileRepo, storageDriver)
	fileHandler := file.NewHandler(fileSvc, storageDriver)

	rl := ratelimit.New(10, 20) // 10 req/s, burst 20

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(rl.Middleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"database": "connected",
		})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/me", authHandler.Me)
		})
	})

	r.Route("/api/folders", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", folderHandler.Create)
		r.Get("/", folderHandler.ListContents)
		r.Get("/starred", folderHandler.ListStarred)
		r.Get("/trash", folderHandler.ListTrash)
		r.Get("/{id}", folderHandler.GetByID)
		r.Get("/{id}/contents", folderHandler.ListContents)
		r.Patch("/{id}", folderHandler.Rename)
		r.Delete("/{id}", folderHandler.Trash)
		r.Post("/{id}/move", folderHandler.Move)
		r.Post("/{id}/restore", folderHandler.Restore)
		r.Post("/{id}/permanent-delete", folderHandler.PermanentDelete)
		r.Post("/{id}/star", folderHandler.ToggleStarred)
	})

	r.Route("/api/files", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", fileHandler.ListContents)
		r.Get("/starred", fileHandler.ListStarred)
		r.Get("/trash", fileHandler.ListTrash)
		r.Post("/upload", fileHandler.Upload)
		r.Get("/{id}", fileHandler.GetByID)
		r.Get("/{id}/preview", fileHandler.Preview)
		r.Get("/{id}/serve", fileHandler.ServeContent)
		r.Get("/{id}/download", fileHandler.Download)
		r.Patch("/{id}", fileHandler.Rename)
		r.Delete("/{id}", fileHandler.Trash)
		r.Post("/{id}/move", fileHandler.Move)
		r.Post("/{id}/restore", fileHandler.Restore)
		r.Post("/{id}/permanent-delete", fileHandler.PermanentDelete)
		r.Post("/{id}/star", fileHandler.ToggleStarred)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server failed to shutdown gracefully", "error", err)
			_ = server.Close()
		}
	}

	slog.Info("server stopped cleanly")
}
