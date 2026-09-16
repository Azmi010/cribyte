package file

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azmi010/cribyte/apps/backend/internal/db"
	"github.com/Azmi010/cribyte/apps/backend/internal/storage"
	"github.com/google/uuid"
)

const (
	MaxUploadSize     = 100 << 20 // 100 MB
	textPreviewLimit  = 1 << 20   // 1 MB
)

var (
	ErrNotFound     = errors.New("file not found")
	ErrNameConflict = errors.New("file with this name already exists")
	ErrTooLarge     = errors.New("file exceeds maximum upload size")
	ErrNotTextFile  = errors.New("file is not a text file")
)

type Service struct {
	repo    *Repository
	storage storage.Storage
}

func NewService(repo *Repository, storage storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

type UploadResult struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	MimeType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	Extension      *string   `json:"extension"`
	ParentFolderID *string   `json:"parent_folder_id"`
	OwnerID        string    `json:"owner_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type FileResult struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	MimeType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	Extension      *string   `json:"extension"`
	Starred        bool      `json:"starred"`
	ParentFolderID *string   `json:"parent_folder_id"`
	OwnerID        string    `json:"owner_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RenameInput struct {
	Name string
}

type MoveInput struct {
	ParentFolderID *string
}

type PreviewKind string

const (
	PreviewKindImage PreviewKind = "image"
	PreviewKindVideo PreviewKind = "video"
	PreviewKindAudio PreviewKind = "audio"
	PreviewKindPDF   PreviewKind = "pdf"
	PreviewKindText  PreviewKind = "text"
	PreviewKindOther PreviewKind = "other"
)

type PreviewResult struct {
	Kind       PreviewKind `json:"kind"`
	MimeType   string      `json:"mime_type"`
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
	Content    *string     `json:"content,omitempty"`
	SignedURL  *string     `json:"signed_url,omitempty"`
}

func toResult(f db.File) FileResult {
	var ext *string
	if f.Extension.Valid {
		ext = &f.Extension.String
	}
	var parentID *string
	if f.ParentFolderID.Valid {
		parentID = &f.ParentFolderID.String
	}
	return FileResult{
		ID:             f.ID,
		Name:           f.Name,
		MimeType:       f.MimeType,
		Size:           f.Size,
		Extension:      ext,
		Starred:        f.Starred,
		ParentFolderID: parentID,
		OwnerID:        f.OwnerID,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

func (s *Service) Upload(ctx context.Context, ownerID string, name string, parentFolderID *string, reader io.Reader, size int64) (UploadResult, error) {
	if size > MaxUploadSize {
		return UploadResult{}, ErrTooLarge
	}

	var parentID sql.NullString
	if parentFolderID != nil {
		parentID = sql.NullString{String: *parentFolderID, Valid: true}
	}

	ext := strings.ToLower(filepath.Ext(name))
	mimeType := detectMimeType(name, ext)

	resolvedName := s.resolveUniqueName(ctx, ownerID, name, parentID)

	storageKey := generateStorageKey(ownerID, resolvedName)

	if err := s.storage.Put(ctx, storageKey, reader, size); err != nil {
		slog.Error("failed to put file to storage", "error", err)
		return UploadResult{}, err
	}

	now := time.Now()
	var extNull sql.NullString
	if ext != "" {
		extNull = sql.NullString{String: ext, Valid: true}
	}

	file, err := s.repo.Create(ctx, db.CreateFileParams{
		ID:             uuid.New().String(),
		Name:           resolvedName,
		MimeType:       mimeType,
		Size:           size,
		StorageKey:     storageKey,
		Extension:      extNull,
		ParentFolderID: parentID,
		OwnerID:        ownerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		slog.Error("failed to create file record", "error", err)
		_ = s.storage.Delete(ctx, storageKey)
		return UploadResult{}, err
	}

	slog.Info("file uploaded", "id", file.ID, "name", file.Name, "size", file.Size)
	return UploadResult{
		ID:             file.ID,
		Name:           file.Name,
		MimeType:       file.MimeType,
		Size:           file.Size,
		Extension:      strPtr(ext),
		ParentFolderID: parentFolderID,
		OwnerID:        file.OwnerID,
		CreatedAt:      file.CreatedAt,
		UpdatedAt:      file.UpdatedAt,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id, ownerID string) (FileResult, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}
	return toResult(file), nil
}

func (s *Service) ListContents(ctx context.Context, ownerID string, parentFolderID *string) ([]FileResult, error) {
	var parentID sql.NullString
	if parentFolderID != nil {
		parentID = sql.NullString{String: *parentFolderID, Valid: true}
	}

	var files []db.File
	var err error
	if parentFolderID == nil {
		files, err = s.repo.ListRoot(ctx, ownerID)
	} else {
		files, err = s.repo.ListByParent(ctx, ownerID, parentID)
	}
	if err != nil {
		slog.Error("failed to list files", "error", err)
		return nil, err
	}

	results := make([]FileResult, len(files))
	for i, f := range files {
		results[i] = toResult(f)
	}
	return results, nil
}

func (s *Service) Download(ctx context.Context, id, ownerID string) (io.ReadCloser, db.File, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return nil, db.File{}, ErrNotFound
	}

	reader, err := s.storage.Get(ctx, file.StorageKey)
	if err != nil {
		slog.Error("failed to get file from storage", "key", file.StorageKey, "error", err)
		return nil, db.File{}, err
	}

	return reader, file, nil
}

func (s *Service) Rename(ctx context.Context, id, ownerID string, input RenameInput) (FileResult, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}

	resolvedName := s.resolveUniqueNameForRename(ctx, ownerID, input.Name, file.ParentFolderID)

	ext := strings.ToLower(filepath.Ext(input.Name))
	var extNull sql.NullString
	if ext != "" {
		extNull = sql.NullString{String: ext, Valid: true}
	}

	updated, err := s.repo.Rename(ctx, db.RenameFileParams{
		Name:      resolvedName,
		Extension: extNull,
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to rename file", "error", err)
		return FileResult{}, err
	}

	return toResult(updated), nil
}

func (s *Service) Move(ctx context.Context, id, ownerID string, input MoveInput) (FileResult, error) {
	_, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}

	var targetParentID sql.NullString
	if input.ParentFolderID != nil {
		targetParentID = sql.NullString{String: *input.ParentFolderID, Valid: true}
	}

	updated, err := s.repo.Move(ctx, db.MoveFileParams{
		ParentFolderID: targetParentID,
		UpdatedAt:      time.Now(),
		ID:             id,
		OwnerID:        ownerID,
	})
	if err != nil {
		slog.Error("failed to move file", "error", err)
		return FileResult{}, err
	}

	return toResult(updated), nil
}

func (s *Service) Trash(ctx context.Context, id, ownerID string) (FileResult, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}

	now := time.Now()
	result, err := s.repo.SoftDelete(ctx, db.SoftDeleteFileParams{
		DeletedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to trash file", "error", err)
		return FileResult{}, err
	}

	_ = file
	slog.Info("file trashed", "id", id)
	return toResult(result), nil
}

func (s *Service) Restore(ctx context.Context, id, ownerID string) (FileResult, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}

	if !file.Deleted {
		return FileResult{}, errors.New("file is not in trash")
	}

	result, err := s.repo.Restore(ctx, db.RestoreFileParams{
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to restore file", "error", err)
		return FileResult{}, err
	}

	slog.Info("file restored", "id", id)
	return toResult(result), nil
}

func (s *Service) PermanentDelete(ctx context.Context, id, ownerID string) error {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return ErrNotFound
	}

	if !file.Deleted {
		return errors.New("file must be in trash before permanent delete")
	}

	if err := s.storage.Delete(ctx, file.StorageKey); err != nil {
		slog.Error("failed to delete file from storage", "key", file.StorageKey, "error", err)
	}

	if err := s.repo.PermanentDelete(ctx, db.PermanentDeleteFileParams{
		ID:      id,
		OwnerID: ownerID,
	}); err != nil {
		slog.Error("failed to permanent delete file", "error", err)
		return err
	}

	slog.Info("file permanently deleted", "id", id)
	return nil
}

func (s *Service) ToggleStarred(ctx context.Context, id, ownerID string, starred bool) (FileResult, error) {
	_, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return FileResult{}, ErrNotFound
	}

	result, err := s.repo.SetStarred(ctx, db.SetFileStarredParams{
		Starred:   starred,
		UpdatedAt: time.Now(),
		ID:        id,
		OwnerID:   ownerID,
	})
	if err != nil {
		slog.Error("failed to toggle starred", "error", err)
		return FileResult{}, err
	}

	return toResult(result), nil
}

func (s *Service) Search(ctx context.Context, ownerID, query string) ([]FileResult, error) {
	pattern := "%" + query + "%"
	files, err := s.repo.SearchByName(ctx, db.SearchFilesByNameParams{
		OwnerID: ownerID,
		Name:    pattern,
	})
	if err != nil {
		slog.Error("failed to search files", "error", err)
		return nil, err
	}

	results := make([]FileResult, len(files))
	for i, f := range files {
		results[i] = toResult(f)
	}
	return results, nil
}

func (s *Service) Preview(ctx context.Context, id, ownerID string) (PreviewResult, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return PreviewResult{}, ErrNotFound
	}

	kind := detectPreviewKind(file.MimeType)

	result := PreviewResult{
		Kind:     kind,
		MimeType: file.MimeType,
		Name:     file.Name,
		Size:     file.Size,
	}

	switch kind {
	case PreviewKindText:
		reader, err := s.storage.Get(ctx, file.StorageKey)
		if err != nil {
			return PreviewResult{}, err
		}
		defer reader.Close()

		limitedReader := io.LimitReader(reader, textPreviewLimit)
		contentBytes, err := io.ReadAll(limitedReader)
		if err != nil {
			return PreviewResult{}, err
		}
		content := string(contentBytes)
		result.Content = &content

	default:
		signedURL, err := s.storage.URL(ctx, file.StorageKey, 1*time.Hour)
		if err != nil {
			return PreviewResult{}, err
		}
		result.SignedURL = &signedURL
	}

	return result, nil
}

func (s *Service) GetStorageKey(ctx context.Context, id, ownerID string) (string, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return "", ErrNotFound
	}
	return file.StorageKey, nil
}

func (s *Service) GetMimeType(ctx context.Context, id, ownerID string) (string, error) {
	file, err := s.repo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return "", ErrNotFound
	}
	return file.MimeType, nil
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

func (s *Service) resolveUniqueNameForRename(ctx context.Context, ownerID, name string, parentFolderID sql.NullString) string {
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

func generateStorageKey(ownerID, name string) string {
	return fmt.Sprintf("%s/%s", ownerID, uuid.New().String()+"_"+name)
}

func detectMimeType(name, ext string) string {
	if ext != "" {
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType
		}
	}

	if ext == "" || ext == "." {
		return "application/octet-stream"
	}

	return "application/octet-stream"
}

func detectPreviewKind(mimeType string) PreviewKind {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return PreviewKindImage
	case strings.HasPrefix(mimeType, "video/"):
		return PreviewKindVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return PreviewKindAudio
	case mimeType == "application/pdf":
		return PreviewKindPDF
	case isTextMimeType(mimeType):
		return PreviewKindText
	default:
		return PreviewKindOther
	}
}

func isTextMimeType(mimeType string) bool {
	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/typescript",
		"application/x-yaml",
		"application/yaml",
	}
	for _, t := range textTypes {
		if strings.HasPrefix(mimeType, t) {
			return true
		}
	}
	return false
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
