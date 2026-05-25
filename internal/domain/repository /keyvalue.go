package repository

import "context"

// KeyValueRepository defines the interface for key-value operations
type KeyValueRepository interface {
	Set(ctx context.Context, key, value string)
	Get(ctx context.Context, key string) (string, bool)
	Del(ctx context.Context, key string) int
	Expire(ctx context.Context, key string, seconds int) bool
	TTL(ctx context.Context, key string) int64
	Persist(ctx context.Context, key string) bool
	StartCleanup(intervalMs int64)
	StopCleanup()
}
