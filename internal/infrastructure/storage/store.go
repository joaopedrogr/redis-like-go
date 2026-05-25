package storage

import (
	"context"
	"sync"
	"time"

	"redis-like-go/internal/domain/entity"
	"redis-like-go/internal/domain/repository "
)

// Store implements KeyValueRepository - thread-safe key-value store with TTL support
type Store struct {
	data        map[string]*entity.Item
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// NewStore creates a new Store instance
func NewStore() repository.KeyValueRepository {
	return &Store{
		data:        make(map[string]*entity.Item),
		stopCleanup: make(chan struct{}),
	}
}

// Set stores or updates a key-value pair
func (s *Store) Set(ctx context.Context, key, value string) {
	if ctx.Err() != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &entity.Item{Value: value, ExpiresAt: nil}
}

// Get retrieves a value by key. Returns the value and true if found and not expired, false otherwise
func (s *Store) Get(ctx context.Context, key string) (string, bool) {
	if ctx.Err() != nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return "", false
	}

	// Check if expired
	if item.IsExpired(time.Now().Unix()) {
		return "", false
	}

	return item.Value, true
}

// Del removes a key and returns the number of keys removed (0 or 1)
func (s *Store) Del(ctx context.Context, key string) int {
	if ctx.Err() != nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return 1
	}
	return 0
}

// Expire sets the TTL for a key in seconds. Returns true if key exists, false otherwise
func (s *Store) Expire(ctx context.Context, key string, seconds int) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	expiresAt := time.Now().Unix() + int64(seconds)
	item.ExpiresAt = &expiresAt
	return true
}

// TTL returns the remaining time-to-live in seconds for a key.
// Returns -1 if key does not exist or has no expiration, or the remaining seconds
func (s *Store) TTL(ctx context.Context, key string) int64 {
	if ctx.Err() != nil {
		return -1
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return -1
	}

	if item.ExpiresAt == nil {
		return -1
	}

	now := time.Now().Unix()
	remaining := *item.ExpiresAt - now

	if remaining <= 0 {
		return -1 // Already expired (will be cleaned up)
	}

	return remaining
}

// Persist removes the expiration from a key. Returns true if key exists, false otherwise
func (s *Store) Persist(ctx context.Context, key string) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	item.ExpiresAt = nil
	return true
}

// StartCleanup starts a background goroutine that periodically removes expired keys
func (s *Store) StartCleanup(intervalMs int64) {
	interval := time.Duration(intervalMs) * time.Millisecond
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.stopCleanup:
				return
			}
		}
	}()
}

// StopCleanup stops the cleanup goroutine
func (s *Store) StopCleanup() {
	close(s.stopCleanup)
}

// cleanupExpired removes all expired keys
func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range s.data {
		if item.IsExpired(now) {
			delete(s.data, key)
		}
	}
}
