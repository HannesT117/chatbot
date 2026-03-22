package session

import (
	"fmt"
	"sync"
)

// SessionStore defines the interface for session persistence.
type SessionStore interface {
	Get(id string) (*Session, error)
	Save(session *Session) error
	Delete(id string) error
}

// InMemoryStore is a thread-safe in-memory implementation of SessionStore.
type InMemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewInMemoryStore returns an initialised InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		sessions: make(map[string]*Session),
	}
}

// Get retrieves a session by ID. Returns an error if the session does not exist.
func (s *InMemoryStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess, nil
}

// Save stores or replaces a session.
func (s *InMemoryStore) Save(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

// Delete removes a session by ID. It is a no-op if the session does not exist.
func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}
