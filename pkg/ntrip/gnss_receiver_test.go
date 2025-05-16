package ntrip

import (
	"errors"
	"io"
	"testing"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStream is a mock implementation of the gnssgo.Stream type
type MockStream struct {
	mock.Mock
	InitStreamCalled bool
	OpenStreamCalled bool
	OpenStreamArgs   []interface{}
	StreamReadCalled bool
	StreamReadArgs   []interface{}
	StreamReadReturn []interface{}
	State            int
	Msg              string
}

func (m *MockStream) InitStream() {
	m.InitStreamCalled = true
}

func (m *MockStream) OpenStream(strtype, strmode int, path string) int {
	m.OpenStreamCalled = true
	m.OpenStreamArgs = []interface{}{strtype, strmode, path}
	args := m.Called(strtype, strmode, path)
	return args.Int(0)
}

func (m *MockStream) StreamRead(buff []byte, n int) int {
	m.StreamReadCalled = true
	m.StreamReadArgs = []interface{}{buff, n}
	args := m.Called(buff, n)
	return args.Int(0)
}

func (m *MockStream) StreamClose() {
	m.Called()
}

// TestNewGNSSReceiver tests the NewGNSSReceiver function
func TestNewGNSSReceiver(t *testing.T) {
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
	
	// Test with valid parameters
	receiver, err := NewGNSSReceiver("COM1:9600:8:N:1")
	
	// Restore original functions
	gnssgo.Stream.InitStream = originalInitStream
	gnssgo.Stream.OpenStream = originalOpenStream
	
	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, receiver)
	assert.Equal(t, "COM1:9600:8:N:1", receiver.port)
	assert.True(t, receiver.open)
}

// TestGNSSReceiverRead tests the Read method
func TestGNSSReceiverRead(t *testing.T) {
	// Create a receiver with a mock stream
	receiver := &GNSSReceiver{
		port: "COM1:9600:8:N:1",
		open: true,
	}
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	receiver.stream = mockStream
	
	// Setup mock expectations
	testData := []byte("test data")
	mockStream.On("StreamRead", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, testData)
	}).Return(len(testData))
	
	// Test read
	buffer := make([]byte, 1024)
	n, err := receiver.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, buffer[:n])
}

// TestGNSSReceiverReadNotOpen tests the Read method when not open
func TestGNSSReceiverReadNotOpen(t *testing.T) {
	// Create a receiver
	receiver := &GNSSReceiver{
		port: "COM1:9600:8:N:1",
		open: false,
	}
	
	// Test read when not open
	buffer := make([]byte, 1024)
	_, err := receiver.Read(buffer)
	assert.Error(t, err)
	assert.Equal(t, "receiver not open", err.Error())
}

// TestGNSSReceiverClose tests the Close method
func TestGNSSReceiverClose(t *testing.T) {
	// Create a receiver with a mock stream
	receiver := &GNSSReceiver{
		port: "COM1:9600:8:N:1",
		open: true,
	}
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	receiver.stream = mockStream
	
	// Setup mock expectations
	mockStream.On("StreamClose").Return()
	
	// Test close
	err := receiver.Close()
	assert.NoError(t, err)
	assert.False(t, receiver.open)
	mockStream.AssertCalled(t, "StreamClose")
}

// TestGNSSReceiverIsOpen tests the IsOpen method
func TestGNSSReceiverIsOpen(t *testing.T) {
	// Create a receiver
	receiver := &GNSSReceiver{
		port: "COM1:9600:8:N:1",
	}
	
	// Test when not open
	assert.False(t, receiver.IsOpen())
	
	// Test when open
	receiver.open = true
	assert.True(t, receiver.IsOpen())
}

// TestGNSSReceiverGetStream tests the GetStream method
func TestGNSSReceiverGetStream(t *testing.T) {
	// Create a receiver
	receiver := &GNSSReceiver{
		port: "COM1:9600:8:N:1",
	}
	
	// Test GetStream
	stream := receiver.GetStream()
	assert.NotNil(t, stream)
	assert.Equal(t, &receiver.stream, stream)
}
