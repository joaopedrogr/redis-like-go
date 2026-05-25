package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"redis-like-go/internal/container"
)

func startTestServer(t *testing.T, enableAOF bool) (string, func()) {
	// Find available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	// Remove AOF file if exists
	if enableAOF {
		os.Remove("test_data.aof")
	}

	// Create container
	ctn, err := container.NewContainer(enableAOF, "test_data.aof")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Replay AOF if enabled
	if enableAOF && ctn.Persistence != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ctn.Persistence.Replay(ctx, ctn.Store); err != nil {
			t.Logf("AOF replay warning: %v", err)
		}
	}

	// Start cleanup
	ctn.Store.StartCleanup(1000) // 1000ms = 1 second

	// Start server
	serverListener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		for {
			conn, err := serverListener.Accept()
			if err != nil {
				return
			}
			go ctn.TCPHandler.HandleConnection(conn)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		ctn.Store.StopCleanup()
		serverListener.Close()
		ctn.Close()
		if enableAOF {
			os.Remove("test_data.aof")
		}
	}

	return addr, cleanup
}

func sendCommand(t *testing.T, conn net.Conn, cmd string) string {
	_, err := fmt.Fprintf(conn, "%s\n", cmd)
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		t.Fatalf("Failed to read response")
	}
	return strings.TrimSpace(scanner.Text())
}

func TestServerSETGET(t *testing.T) {
	addr, cleanup := startTestServer(t, false)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// SET
	response := sendCommand(t, conn, "SET key1 value1")
	if response != "OK" {
		t.Errorf("Expected OK, got %s", response)
	}

	// GET
	response = sendCommand(t, conn, "GET key1")
	if response != "value1" {
		t.Errorf("Expected 'value1', got '%s'", response)
	}

	// GET nonexistent
	response = sendCommand(t, conn, "GET nonexistent")
	if response != "nil" {
		t.Errorf("Expected 'nil', got '%s'", response)
	}
}

func TestServerDEL(t *testing.T) {
	addr, cleanup := startTestServer(t, false)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	sendCommand(t, conn, "SET key1 value1")
	sendCommand(t, conn, "SET key2 value2")

	// DEL one key
	response := sendCommand(t, conn, "DEL key1")
	if response != "1" {
		t.Errorf("Expected '1', got %s", response)
	}

	// GET deleted key
	response = sendCommand(t, conn, "GET key1")
	if response != "nil" {
		t.Errorf("Expected 'nil', got '%s'", response)
	}

	// DEL nonexistent
	response = sendCommand(t, conn, "DEL nonexistent")
	if response != "0" {
		t.Errorf("Expected '0', got %s", response)
	}
}

func TestServerEXPIRE(t *testing.T) {
	addr, cleanup := startTestServer(t, false)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	sendCommand(t, conn, "SET key1 value1")

	// EXPIRE
	response := sendCommand(t, conn, "EXPIRE key1 1")
	if response != "OK" {
		t.Errorf("Expected 'OK', got %s", response)
	}

	// TTL
	response = sendCommand(t, conn, "TTL key1")
	ttl, err := strconv.Atoi(response)
	if err != nil || ttl < 0 || ttl > 1 {
		t.Errorf("Expected TTL between 0 and 1, got %s", response)
	}

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	// GET expired key
	response = sendCommand(t, conn, "GET key1")
	if response != "nil" {
		t.Errorf("Expected 'nil' for expired key, got %s", response)
	}
}

func TestServerPERSIST(t *testing.T) {
	addr, cleanup := startTestServer(t, false)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	sendCommand(t, conn, "SET key1 value1")
	sendCommand(t, conn, "EXPIRE key1 60")

	// PERSIST
	response := sendCommand(t, conn, "PERSIST key1")
	if response != "OK" {
		t.Errorf("Expected 'OK', got %s", response)
	}

	// TTL should be -1
	response = sendCommand(t, conn, "TTL key1")
	if response != "-1" {
		t.Errorf("Expected '-1', got %s", response)
	}
}

func TestServerAOF(t *testing.T) {
	addr, cleanup := startTestServer(t, true)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// SET some data
	sendCommand(t, conn, "SET key1 value1")
	sendCommand(t, conn, "SET key2 value2")
	sendCommand(t, conn, "EXPIRE key1 60")
	conn.Close()

	// Restart server (simulate by replaying AOF)
	ctx := context.Background()
	ctn, err := container.NewContainer(true, "test_data.aof")
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer ctn.Close()

	if err := ctn.Persistence.Replay(ctx, ctn.Store); err != nil {
		t.Fatalf("Failed to replay AOF: %v", err)
	}

	// Verify data was restored
	val, found := ctn.Store.Get(ctx, "key1")
	if !found {
		t.Error("key1 should exist after AOF replay")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%s'", val)
	}

	val, found = ctn.Store.Get(ctx, "key2")
	if !found {
		t.Error("key2 should exist after AOF replay")
	}
	if val != "value2" {
		t.Errorf("Expected 'value2', got '%s'", val)
	}
}

func TestServerInvalidCommand(t *testing.T) {
	addr, cleanup := startTestServer(t, false)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	response := sendCommand(t, conn, "INVALID command")
	if !strings.HasPrefix(response, "ERR") {
		t.Errorf("Expected error response, got '%s'", response)
	}
}
