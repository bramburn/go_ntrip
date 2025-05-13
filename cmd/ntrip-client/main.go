package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bramburn/go_ntrip/internal/ntrip"
	"github.com/bramburn/go_ntrip/internal/position"
)

func main() {
	// Parse command line flags
	address := flag.String("address", "", "NTRIP server address (e.g., 192.168.0.64)")
	port := flag.String("port", "2101", "NTRIP server port")
	username := flag.String("user", "", "Username for NTRIP server")
	password := flag.String("pass", "", "Password for NTRIP server")
	mountpoint := flag.String("mount", "", "Mountpoint name")
	outputFile := flag.String("output", "", "Output file path (default: ./base_position.json)")
	timeout := flag.Duration("timeout", 60*time.Second, "Timeout for connection")
	flag.Parse()

	// Check required parameters
	if *address == "" {
		fmt.Println("Error: NTRIP server address is required")
		flag.Usage()
		os.Exit(1)
	}

	if *mountpoint == "" {
		fmt.Println("Error: Mountpoint is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output file if not specified
	if *outputFile == "" {
		execPath, err := os.Executable()
		if err != nil {
			execPath = "."
		}
		*outputFile = filepath.Join(filepath.Dir(execPath), "base_position.json")
	}

	// Construct URL
	url := fmt.Sprintf("http://%s:%s", *address, *port)

	// Create NTRIP client
	client := ntrip.NewClient(url, *username, *password, *mountpoint)

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived shutdown signal")
		cancel()
	}()

	// Connect to NTRIP server
	fmt.Printf("Connecting to NTRIP server at %s...\n", url)
	stream, err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error connecting to NTRIP server: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	fmt.Println("Connected to NTRIP server.")
	fmt.Println("Waiting for position data...")

	// Read RTCM data
	buffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Timeout or cancellation")
			return
		default:
			n, err := stream.Read(buffer)
			if err != nil {
				fmt.Printf("Error reading from NTRIP stream: %v\n", err)
				return
			}
			if n > 0 {
				// Process RTCM data
				fmt.Printf("Received %d bytes of RTCM data\r", n)
			}
		}
	}
}

// savePosition saves the position to a JSON file
func savePosition(pos *position.Position, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(pos, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
