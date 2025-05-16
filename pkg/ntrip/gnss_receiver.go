package ntrip

import (
	"fmt"
	"io"
	"sync"

	"github.com/bramburn/gnssgo"
)

// GNSSReceiver represents a physical GNSS receiver
type GNSSReceiver struct {
	port   string
	stream gnssgo.Stream
	mutex  sync.Mutex
	open   bool
}

// NewGNSSReceiver creates a new GNSS receiver
func NewGNSSReceiver(port string) (*GNSSReceiver, error) {
	receiver := &GNSSReceiver{
		port: port,
	}

	// Initialize the stream
	receiver.stream.InitStream()

	// Open the serial port
	result := receiver.stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_R, port)

	// Check if the stream was opened successfully
	if result <= 0 || receiver.stream.State <= 0 {
		return nil, fmt.Errorf("failed to connect to GNSS receiver: %s", receiver.stream.Msg)
	}

	receiver.open = true
	return receiver, nil
}

// Read reads data from the GNSS receiver
func (r *GNSSReceiver) Read(p []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.open {
		return 0, fmt.Errorf("receiver not open")
	}

	n := r.stream.StreamRead(p, len(p))
	if n <= 0 {
		return 0, io.EOF
	}

	return n, nil
}

// Close closes the GNSS receiver
func (r *GNSSReceiver) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.open {
		return nil
	}

	r.stream.StreamClose()
	r.open = false
	return nil
}

// IsOpen returns true if the receiver is open
func (r *GNSSReceiver) IsOpen() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.open
}

// GetStream returns the underlying stream
func (r *GNSSReceiver) GetStream() *gnssgo.Stream {
	return &r.stream
}
