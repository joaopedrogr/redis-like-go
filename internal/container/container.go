package container

import (
	"redis-like-go/internal/adapter/handler"
	"redis-like-go/internal/adapter/protocol"
	"redis-like-go/internal/domain/repository"
	"redis-like-go/internal/infrastructure/persistence"
	"redis-like-go/internal/infrastructure/storage"
	"redis-like-go/internal/usecase"
)

// Container holds all dependencies
type Container struct {
	Store          repository.KeyValueRepository
	Persistence    repository.PersistenceRepository
	CommandHandler *usecase.CommandHandler
	TCPHandler     *handler.TCPHandler
	Parser         *protocol.Parser
}

// NewContainer creates a new dependency injection container
func NewContainer(enableAOF bool, aofFilepath string) (*Container, error) {
	// Create store (infrastructure layer)
	store := storage.NewStore()

	// Create persistence if enabled
	var persistRepo repository.PersistenceRepository
	if enableAOF {
		var err error
		persistRepo, err = persistence.NewAOF(aofFilepath)
		if err != nil {
			return nil, err
		}
	}

	// Create parser
	parser := protocol.NewParser()

	// Create command handler (use case layer)
	commandHandler := usecase.NewCommandHandler(store, persistRepo)

	// Create TCP handler (adapter layer)
	tcpHandler := handler.NewTCPHandler(commandHandler)

	return &Container{
		Store:          store,
		Persistence:    persistRepo,
		CommandHandler: commandHandler,
		TCPHandler:     tcpHandler,
		Parser:         parser,
	}, nil
}

// Close closes all resources that need cleanup
func (c *Container) Close() error {
	if c.Persistence != nil {
		return c.Persistence.Close()
	}
	return nil
}
