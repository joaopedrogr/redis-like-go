package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <host:port>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s localhost:6379\n", os.Args[0])
		os.Exit(1)
	}

	addr := os.Args[1]

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", addr)
	fmt.Println("Type commands (or 'QUIT' to exit):")

	// Read responses in a goroutine
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			response := scanner.Text()
			fmt.Println(response)
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		}
	}()

	// Read commands from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Send command
		_, err := fmt.Fprintf(conn, "%s\n", line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending command: %v\n", err)
			break
		}

		// Check for QUIT
		parts := strings.Fields(line)
		if len(parts) > 0 && strings.ToUpper(parts[0]) == "QUIT" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
