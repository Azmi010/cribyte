package session

import (
	"sync"
	"time"
)

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemoryStore) Create(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	session, exists := s.sessions[id]
	s.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	if time.Now().After(session.ExpiresAt) {
		s.Delete(id)
		return nil, nil
	}

	return session, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}

func (s *MemoryStore) DeleteByUserID(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}

	return nil
}

func (s *MemoryStore) Cleanup() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var count int64

	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
			count++
		}
	}

	return count, nil
}
