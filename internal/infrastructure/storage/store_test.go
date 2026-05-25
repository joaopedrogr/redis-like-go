package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	store.Set(ctx, "key1", "value1")
	store.Set(ctx, "key2", "value2")

	val, found := store.Get(ctx, "key1")
	if !found || val != "value1" {
		t.Errorf("Expected 'value1', got '%s' (found: %v)", val, found)
	}

	val, found = store.Get(ctx, "key2")
	if !found || val != "value2" {
		t.Errorf("Expected 'value2', got '%s' (found: %v)", val, found)
	}

	val, found = store.Get(ctx, "nonexistent")
	if found {
		t.Errorf("Expected key to not exist, but it was found")
	}
}

func TestDel(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	store.Set(ctx, "key1", "value1")
	store.Set(ctx, "key2", "value2")

	count := store.Del(ctx, "key1")
	if count != 1 {
		t.Errorf("Expected 1 deletion, got %d", count)
	}

	count = store.Del(ctx, "key1")
	if count != 0 {
		t.Errorf("Expected 0 deletions, got %d", count)
	}

	_, found := store.Get(ctx, "key1")
	if found {
		t.Errorf("Key should have been deleted")
	}

	_, found = store.Get(ctx, "key2")
	if !found {
		t.Errorf("Key2 should still exist")
	}
}

func TestExpire(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	store.Set(ctx, "key1", "value1")

	success := store.Expire(ctx, "key1", 1)
	if !success {
		t.Errorf("Expected Expire to succeed")
	}

	success = store.Expire(ctx, "nonexistent", 10)
	if success {
		t.Errorf("Expected Expire to fail for nonexistent key")
	}
}

func TestTTL(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	store.Set(ctx, "key1", "value1")
	ttl := store.TTL(ctx, "key1")
	if ttl != -1 {
		t.Errorf("Expected TTL to be -1 (no expiration), got %d", ttl)
	}

	store.Expire(ctx, "key1", 10)
	ttl = store.TTL(ctx, "key1")
	if ttl < 1 || ttl > 10 {
		t.Errorf("Expected TTL between 1 and 10, got %d", ttl)
	}

	ttl = store.TTL(ctx, "nonexistent")
	if ttl != -1 {
		t.Errorf("Expected TTL to be -1 for nonexistent key, got %d", ttl)
	}
}

func TestPersist(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	store.Set(ctx, "key1", "value1")
	store.Expire(ctx, "key1", 10)

	success := store.Persist(ctx, "key1")
	if !success {
		t.Errorf("Expected Persist to succeed")
	}

	ttl := store.TTL(ctx, "key1")
	if ttl != -1 {
		t.Errorf("Expected TTL to be -1 after Persist, got %d", ttl)
	}

	success = store.Persist(ctx, "nonexistent")
	if success {
		t.Errorf("Expected Persist to fail for nonexistent key")
	}
}

func TestExpirationCleanup(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	store.StartCleanup(100) // 100ms
	defer store.StopCleanup()

	store.Set(ctx, "key1", "value1")
	store.Expire(ctx, "key1", 1) // 1 second

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	_, found := store.Get(ctx, "key1")
	if found {
		t.Errorf("Key should have been expired and removed")
	}
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	var wg sync.WaitGroup
	numGoroutines := 100
	opsPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				store.Set(ctx, key, "value")
			}
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				_, found := store.Get(ctx, key)
				if !found {
					t.Errorf("Key %s should exist", key)
				}
			}
		}(i)
	}

	wg.Wait()
}
