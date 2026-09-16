package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/Azmi010/my-drive/apps/backend/internal/config"
	"github.com/Azmi010/my-drive/apps/backend/internal/db"
	"github.com/Azmi010/my-drive/apps/backend/internal/session"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists       = errors.New("email already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
)

type Service struct {
	repo     *Repository
	sessions session.SessionStore
	cfg      *config.Config
}

func NewService(repo *Repository, sessions session.SessionStore, cfg *config.Config) *Service {
	return &Service{repo: repo, sessions: sessions, cfg: cfg}
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type UserResult struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
}

type LoginResult struct {
	User      UserResult
	SessionID string
	MaxAge    int
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (UserResult, error) {
	email := normalizeEmail(input.Email)

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return UserResult{}, err
	}

	now := time.Now()
	user, err := s.repo.Create(ctx, db.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         input.Name,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return UserResult{}, ErrEmailExists
		}
		slog.Error("failed to create user", "error", err)
		return UserResult{}, err
	}

	slog.Info("user registered", "id", user.ID, "email", user.Email)
	return UserResult{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := normalizeEmail(input.Email)

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredential
	}

	sessionID, err := generateSessionID()
	if err != nil {
		slog.Error("failed to generate session ID", "error", err)
		return LoginResult{}, err
	}

	sess := &session.Session{
		ID:        sessionID,
		UserID:    user.ID,
		UserEmail: user.Email,
		UserName:  user.Name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.SessionExpiryHours) * time.Hour),
	}

	if err := s.sessions.Create(sess); err != nil {
		slog.Error("failed to create session", "error", err)
		return LoginResult{}, err
	}

	slog.Info("user logged in", "id", user.ID, "email", user.Email)
	return LoginResult{
		User: UserResult{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		SessionID: sessionID,
		MaxAge:    s.cfg.SessionExpiryHours * 3600,
	}, nil
}

func (s *Service) Logout(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.sessions.Delete(sessionID)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (db.User, error) {
	return s.repo.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
