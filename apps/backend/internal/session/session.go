package session

import "time"

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionStore interface {
	Create(session *Session) error
	Get(id string) (*Session, error)
	Delete(id string) error
	DeleteByUserID(userID string) error
	Cleanup() (int64, error)
}
