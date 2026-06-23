package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Store is an in-memory, TTL-bounded store keyed by an opaque random ID. It is
// domain-agnostic: callers decide what type T to keep in it. Safe for
// concurrent use. Suitable for a single-instance deployment.
//
// Expired entries are reclaimed two ways: lazily on Get, and proactively by a
// background janitor so abandoned entries (never accessed again) cannot pile up.
type Store[T any] struct {
	mu    sync.RWMutex
	items map[string]item[T]
	ttl   time.Duration
	stop  chan struct{}
}

type item[T any] struct {
	value     T
	createdAt time.Time
}

func NewStore[T any](ttl time.Duration) *Store[T] {
	s := &Store[T]{
		items: make(map[string]item[T]),
		ttl:   ttl,
		stop:  make(chan struct{}),
	}
	if ttl > 0 {
		go s.janitor()
	}
	return s
}

// janitor periodically evicts expired entries until the store is closed. The
// sweep interval is the TTL itself, so an abandoned entry lives at most ~2×TTL.
func (s *Store[T]) janitor() {
	t := time.NewTicker(s.ttl)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.evictExpired()
		case <-s.stop:
			return
		}
	}
}

func (s *Store[T]) evictExpired() {
	now := time.Now()
	s.mu.Lock()
	for id, it := range s.items {
		if now.Sub(it.createdAt) > s.ttl {
			delete(s.items, id)
		}
	}
	s.mu.Unlock()
}

// Close stops the background janitor. Safe to call once; intended for tests and
// graceful shutdown. Not required for process exit.
func (s *Store[T]) Close() {
	close(s.stop)
}

// Create stores value under a fresh random ID and returns the ID.
func (s *Store[T]) Create(value T) (string, error) {
	id, err := randomID()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.items[id] = item[T]{value: value, createdAt: time.Now()}
	s.mu.Unlock()
	return id, nil
}

// Get returns the stored value, or false if it is missing or expired.
func (s *Store[T]) Get(id string) (T, bool) {
	s.mu.RLock()
	it, ok := s.items[id]
	s.mu.RUnlock()
	if !ok {
		var zero T
		return zero, false
	}
	if s.ttl > 0 && time.Since(it.createdAt) > s.ttl {
		s.Delete(id)
		var zero T
		return zero, false
	}
	return it.value, true
}

func (s *Store[T]) Delete(id string) {
	s.mu.Lock()
	delete(s.items, id)
	s.mu.Unlock()
}

func randomID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
