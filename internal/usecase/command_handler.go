package usecase

import (
	"context"
	"fmt"
	"redis-like-go/internal/domain/repository "
	"strconv"
	"strings"

	"redis-like-go/internal/adapter/protocol"
	"redis-like-go/internal/domain/command"
)

// CommandHandler handles command execution
type CommandHandler struct {
	store   repository.KeyValueRepository
	persist repository.PersistenceRepository
	parser  *protocol.Parser
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(
	store repository.KeyValueRepository,
	persist repository.PersistenceRepository,
) *CommandHandler {
	return &CommandHandler{
		store:   store,
		persist: persist,
		parser:  protocol.NewParser(),
	}
}

// ExecuteCommand executes a command and returns the response
func (h *CommandHandler) ExecuteCommand(ctx context.Context, cmd *protocol.Command) string {
	switch cmd.Type {
	case command.SET:
		return h.handleSet(ctx, cmd.Args)
	case command.GET:
		return h.handleGet(ctx, cmd.Args)
	case command.DEL:
		return h.handleDel(ctx, cmd.Args)
	case command.EXPIRE:
		return h.handleExpire(ctx, cmd.Args)
	case command.TTL:
		return h.handleTTL(ctx, cmd.Args)
	case command.PERSIST:
		return h.handlePersist(ctx, cmd.Args)
	default:
		return h.parser.FormatError(fmt.Sprintf("unknown command: %s", cmd.Type))
	}
}

func (h *CommandHandler) handleSet(ctx context.Context, args []string) string {
	if len(args) < 2 {
		return h.parser.FormatError("SET requires at least 2 arguments")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")
	h.store.Set(ctx, key, value)

	if h.persist != nil {
		h.persist.Append(ctx, command.SET.String(), args)
	}

	return h.parser.FormatOK()
}

func (h *CommandHandler) handleGet(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("GET requires 1 argument")
	}

	value, found := h.store.Get(ctx, args[0])
	if found {
		return value
	}
	return h.parser.FormatNil()
}

func (h *CommandHandler) handleDel(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("DEL requires at least 1 argument")
	}

	count := 0
	for _, key := range args {
		count += h.store.Del(ctx, key)
	}

	if h.persist != nil && count > 0 {
		for _, key := range args {
			h.persist.Append(ctx, command.DEL.String(), []string{key})
		}
	}

	return h.parser.FormatResponse(count)
}

func (h *CommandHandler) handleExpire(ctx context.Context, args []string) string {
	if len(args) < 2 {
		return h.parser.FormatError("EXPIRE requires 2 arguments")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return h.parser.FormatError("invalid seconds value")
	}

	success := h.store.Expire(ctx, key, seconds)
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.EXPIRE.String(), args)
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(0)
}

func (h *CommandHandler) handleTTL(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("TTL requires 1 argument")
	}

	ttl := h.store.TTL(ctx, args[0])
	return h.parser.FormatResponse(ttl)
}

func (h *CommandHandler) handlePersist(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("PERSIST requires 1 argument")
	}

	success := h.store.Persist(ctx, args[0])
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.PERSIST.String(), args)
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(0)
}
