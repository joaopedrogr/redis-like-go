package repository

import "context"

// PersistenceRepository defines the interface for persistence operations
type PersistenceRepository interface {
	Append(ctx context.Context, command string, args []string) error
	Replay(ctx context.Context, store KeyValueRepository) error
	Close() error
}
