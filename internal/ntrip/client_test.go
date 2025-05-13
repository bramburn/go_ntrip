package ntrip

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	url := "http://example.com"
	username := "user"
	password := "pass"
	mountpoint := "MOUNT"

	client := NewClient(url, username, password, mountpoint)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.URL != url {
		t.Errorf("Expected URL %s, got %s", url, client.URL)
	}

	if client.Username != username {
		t.Errorf("Expected username %s, got %s", username, client.Username)
	}

	if client.Password != password {
		t.Errorf("Expected password %s, got %s", password, client.Password)
	}

	if client.Mountpoint != mountpoint {
		t.Errorf("Expected mountpoint %s, got %s", mountpoint, client.Mountpoint)
	}

	if client.httpClient == nil {
		t.Error("httpClient should be initialized")
	}
}

func TestConnect(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/MOUNT" {
			t.Errorf("Expected path /MOUNT, got %s", r.URL.Path)
		}

		// Check authentication
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected basic authentication")
		}
		if username != "user" {
			t.Errorf("Expected username 'user', got '%s'", username)
		}
		if password != "pass" {
			t.Errorf("Expected password 'pass', got '%s'", password)
		}

		// Send response
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("RTCM data"))
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	// Connect
	ctx := context.Background()
	stream, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stream.Close()

	// Read data
	data, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("Error reading from stream: %v", err)
	}

	// Check data
	if string(data) != "RTCM data" {
		t.Errorf("Expected 'RTCM data', got '%s'", string(data))
	}
}

func TestConnectError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	// Connect
	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetSourcetable(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Send response
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SOURCETABLE 200 OK\r\n" +
			"STR;MOUNT1;Server 1;RTCM 3;1005,1077,1087,1097,1127;2;GPS+GLO+GAL+BDS;SNIP;CHN;31.22;121.46;1;1;SNIP;none;B;N;0;\r\n" +
			"STR;MOUNT2;Server 2;RTCM 3;1005,1077,1087,1097,1127;2;GPS+GLO+GAL+BDS;SNIP;CHN;31.22;121.46;1;1;SNIP;none;B;N;0;\r\n" +
			"ENDSOURCETABLE\r\n"))
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	// Get sourcetable
	ctx := context.Background()
	sourcetable, err := client.GetSourcetable(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check sourcetable
	if sourcetable == nil {
		t.Fatal("Expected non-nil sourcetable")
	}

	if len(sourcetable.Mounts) != 2 {
		t.Errorf("Expected 2 mounts, got %d", len(sourcetable.Mounts))
	}

	if sourcetable.Mounts[0].Name != "MOUNT1" {
		t.Errorf("Expected mount name 'MOUNT1', got '%s'", sourcetable.Mounts[0].Name)
	}

	if sourcetable.Mounts[1].Name != "MOUNT2" {
		t.Errorf("Expected mount name 'MOUNT2', got '%s'", sourcetable.Mounts[1].Name)
	}
}

func TestGetSourcetableError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	// Get sourcetable
	ctx := context.Background()
	_, err := client.GetSourcetable(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestConnectWithMountpoint(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path
		if r.URL.Path != "/MOUNT" {
			t.Errorf("Expected path /MOUNT, got %s", r.URL.Path)
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("RTCM data"))
	}))
	defer server.Close()

	// Test with URL that doesn't include mountpoint
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test with URL that already includes mountpoint
	client = NewClient(server.URL+"/MOUNT", "user", "pass", "")

	_, err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test with URL that has trailing slash
	client = NewClient(server.URL+"/", "user", "pass", "MOUNT")

	_, err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestConnectTimeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "user", "pass", "MOUNT")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Connect should timeout
	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected deadline exceeded error, got: %v", err)
	}
}
