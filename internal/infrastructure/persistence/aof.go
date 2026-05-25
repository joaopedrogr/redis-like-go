package persistence

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"redis-like-go/internal/domain/repository "
)

// AOF implements PersistenceRepository - append-only file persistence
type AOF struct {
	filepath string
	file     *os.File
	mu       sync.Mutex
}

// NewAOF creates a new AOF persistence instance
func NewAOF(filepath string) (repository.PersistenceRepository, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &AOF{
		filepath: filepath,
		file:     file,
	}, nil
}

// Append writes a command to the AOF file
func (a *AOF) Append(ctx context.Context, command string, args []string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	line := command
	if len(args) > 0 {
		line = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}
	line += "\n"

	_, err := a.file.WriteString(line)
	if err != nil {
		return err
	}

	return a.file.Sync()
}

// Replay reads the AOF file and replays all commands to rebuild the store state
func (a *AOF) Replay(ctx context.Context, store repository.KeyValueRepository) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	file, err := os.Open(a.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to replay
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		switch cmd {
		case "SET":
			if len(args) < 2 {
				continue
			}
			key := args[0]
			value := strings.Join(args[1:], " ")
			store.Set(ctx, key, value)

		case "EXPIRE":
			if len(args) < 2 {
				continue
			}
			key := args[0]
			seconds, err := strconv.Atoi(args[1])
			if err != nil {
				continue
			}
			store.Expire(ctx, key, seconds)

		case "PERSIST":
			if len(args) < 1 {
				continue
			}
			store.Persist(ctx, args[0])

		case "DEL":
			if len(args) < 1 {
				continue
			}
			store.Del(ctx, args[0])

		default:
			// Ignore unknown commands
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading AOF file: %w", err)
	}

	return nil
}

// Close closes the AOF file
func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		return a.file.Close()
	}
	return nil
}
