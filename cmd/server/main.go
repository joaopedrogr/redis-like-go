package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"redis-like-go/internal/container"
)

var (
	port      = flag.String("port", "6379", "Port to listen on")
	enableAOF = flag.Bool("aof", false, "Enable append-only file persistence")
)

func main() {
	flag.Parse()

	// Create dependency injection container
	ctn, err := container.NewContainer(*enableAOF, "data.aof")
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	defer ctn.Close()

	// Replay AOF if enabled
	if *enableAOF && ctn.Persistence != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctn.Persistence.Replay(ctx, ctn.Store); err != nil {
			log.Printf("Warning: failed to replay AOF: %v", err)
		} else {
			log.Println("AOF replay completed")
		}
	}

	// Start cleanup goroutine
	ctn.Store.StartCleanup(1000) // 1000ms = 1 second
	defer ctn.Store.StopCleanup()

	// Start TCP server
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", *port, err)
	}
	defer listener.Close()

	log.Printf("Server listening on port %s (AOF: %v)", *port, *enableAOF)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		listener.Close()
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			break
		}

		go ctn.TCPHandler.HandleConnection(conn)
	}
}
