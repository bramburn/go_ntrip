package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bramburn/go_ntrip/internal/ntrip"
	"github.com/bramburn/go_ntrip/internal/rtk"
)

func TestNTRIPClientRTKIntegration(t *testing.T) {
	// Create a test server that sends RTCM data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/MOUNT" {
			t.Errorf("Expected path /MOUNT, got %s", r.URL.Path)
		}

		// Send response
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		// Send 10 chunks of RTCM data
		for i := 0; i < 10; i++ {
			// Create RTCM data
			rtcmData := make([]byte, 1024)
			for j := range rtcmData {
				rtcmData[j] = byte((i*1024 + j) % 256)
			}

			// Write data
			w.Write(rtcmData)
			w.(http.Flusher).Flush()

			// Sleep to simulate real-time data
			time.Sleep(100 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create NTRIP client
	client := ntrip.NewClient(server.URL, "user", "pass", "MOUNT")

	// Create RTK processor
	processor := rtk.NewProcessor()

	// Start processing
	processor.StartProcessing()
	defer processor.StopProcessing()

	// Get solution channel
	solutionChan := processor.GetSolutionChannel()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to NTRIP server
	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stream.Close()

	// Create a channel to signal completion
	doneChan := make(chan struct{})

	// Start goroutine to read from NTRIP stream and process RTCM data
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stream.Read(buffer)
				if err != nil {
					t.Logf("Error reading from NTRIP stream: %v", err)
					doneChan <- struct{}{}
					return
				}
				if n > 0 {
					// Process RTCM data
					processor.ProcessRTCM(buffer[:n])
				}
			}
		}
	}()

	// Wait for solutions
	solutionCount := 0
	timeout := time.After(5 * time.Second)

	for solutionCount < 3 {
		select {
		case <-solutionChan:
			solutionCount++
		case <-doneChan:
			if solutionCount == 0 {
				t.Error("No solutions received before stream ended")
			}
			return
		case <-timeout:
			if solutionCount == 0 {
				t.Error("No solutions received within timeout")
			}
			return
		}
	}

	// We got at least 3 solutions, test passes
}
