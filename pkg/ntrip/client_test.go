package ntrip

import (
	"errors"
	"io"
	"testing"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStream is already defined in gnss_receiver_test.go

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	// Test with valid parameters
	client, err := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "example.com", client.server)
	assert.Equal(t, "2101", client.port)
	assert.Equal(t, "user", client.username)
	assert.Equal(t, "pass", client.password)
	assert.Equal(t, "MOUNT", client.mountpoint)
	assert.False(t, client.connected)
}

// TestClientConnect tests the Connect method
func TestClientConnect(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Create a mock stream
	originalStream := gnssgo.Stream{}
	
	// Save the original functions to restore later
	originalInitStream := originalStream.InitStream
	originalOpenStream := originalStream.OpenStream
	
	// Mock the InitStream and OpenStream functions
	gnssgo.Stream.InitStream = func(s *gnssgo.Stream) {
		// Do nothing
	}
	
	gnssgo.Stream.OpenStream = func(s *gnssgo.Stream, strtype, strmode int, path string) int {
		s.State = 1 // Set state to connected
		return 1    // Return success
	}
	
	// Test connect
	err := client.Connect()
	
	// Restore original functions
	gnssgo.Stream.InitStream = originalInitStream
	gnssgo.Stream.OpenStream = originalOpenStream
	
	// Verify results
	assert.NoError(t, err)
	assert.True(t, client.connected)
}

// TestClientConnectFailure tests the Connect method with a failure
func TestClientConnectFailure(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Create a mock stream
	originalStream := gnssgo.Stream{}
	
	// Save the original functions to restore later
	originalInitStream := originalStream.InitStream
	originalOpenStream := originalStream.OpenStream
	
	// Mock the InitStream and OpenStream functions
	gnssgo.Stream.InitStream = func(s *gnssgo.Stream) {
		// Do nothing
	}
	
	gnssgo.Stream.OpenStream = func(s *gnssgo.Stream, strtype, strmode int, path string) int {
		s.State = 0 // Set state to not connected
		s.Msg = "connection failed"
		return 0    // Return failure
	}
	
	// Test connect
	err := client.Connect()
	
	// Restore original functions
	gnssgo.Stream.InitStream = originalInitStream
	gnssgo.Stream.OpenStream = originalOpenStream
	
	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
	assert.False(t, client.connected)
}

// TestClientDisconnect tests the Disconnect method
func TestClientDisconnect(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations
	mockStream.On("StreamClose").Return()
	
	// Set connected state
	client.connected = true
	
	// Test disconnect
	err := client.Disconnect()
	assert.NoError(t, err)
	assert.False(t, client.connected)
	mockStream.AssertCalled(t, "StreamClose")
}

// TestClientRead tests the Read method
func TestClientRead(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations
	testData := []byte("test data")
	mockStream.On("StreamRead", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, testData)
	}).Return(len(testData))
	
	// Set connected state
	client.connected = true
	
	// Test read
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, buffer[:n])
}

// TestClientReadNotConnected tests the Read method when not connected
func TestClientReadNotConnected(t *testing.T) {
	// Create a client
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Test read when not connected
	buffer := make([]byte, 1024)
	_, err := client.Read(buffer)
	assert.Error(t, err)
	assert.Equal(t, "not connected", err.Error())
}

// TestClientIsConnected tests the IsConnected method
func TestClientIsConnected(t *testing.T) {
	// Create a client
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Test when not connected
	assert.False(t, client.IsConnected())
	
	// Test when connected
	client.connected = true
	assert.True(t, client.IsConnected())
}

// TestClientGetStream tests the GetStream method
func TestClientGetStream(t *testing.T) {
	// Create a client
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Test GetStream
	stream := client.GetStream()
	assert.NotNil(t, stream)
	assert.Equal(t, &client.stream, stream)
}
