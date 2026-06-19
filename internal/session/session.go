package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"RegistryUI/internal/registry"
)

// CookieName is the cookie that carries the opaque session token.
const CookieName = "rui_session"

// Session binds a logged-in user to a configured registry client.
type Session struct {
	Token       string
	RegistryURL string
	Username    string
	Client      *registry.Client
	CreatedAt   time.Time
}

// Store is an in-memory session store. Suitable for a single-instance local tool.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

// NewStore creates a session store. Sessions older than ttl are evicted on access.
func NewStore(ttl time.Duration) *Store {
	return &Store{sessions: make(map[string]*Session), ttl: ttl}
}

// Create registers a new session for the given registry client.
func (s *Store) Create(client *registry.Client, registryURL, username string) (*Session, error) {
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	sess := &Session{
		Token:       token,
		RegistryURL: registryURL,
		Username:    username,
		Client:      client,
		CreatedAt:   time.Now(),
	}
	s.mu.Lock()
	s.sessions[token] = sess
	s.mu.Unlock()
	return sess, nil
}

// Get returns the session for a token, or false if missing/expired.
func (s *Store) Get(token string) (*Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if s.ttl > 0 && time.Since(sess.CreatedAt) > s.ttl {
		s.Delete(token)
		return nil, false
	}
	return sess, true
}

// Delete removes a session.
func (s *Store) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
