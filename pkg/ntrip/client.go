package ntrip

import (
	"fmt"
	"io"
	"sync"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// Client represents an NTRIP client
type Client struct {
	server     string
	port       string
	username   string
	password   string
	mountpoint string
	stream     gnssgo.Stream
	mutex      sync.Mutex
	connected  bool
}

// NewClient creates a new NTRIP client
func NewClient(server, port, username, password, mountpoint string) (*Client, error) {
	return &Client{
		server:     server,
		port:       port,
		username:   username,
		password:   password,
		mountpoint: mountpoint,
	}, nil
}

// Connect connects to the NTRIP server
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	// Initialize the stream
	c.stream.InitStream()

	// Construct the NTRIP path
	ntripPath := fmt.Sprintf("%s:%s@%s:%s/%s",
		c.username, c.password, c.server, c.port, c.mountpoint)

	// Open the NTRIP client stream
	result := c.stream.OpenStream(gnssgo.STR_NTRIPCLI, gnssgo.STR_MODE_R, ntripPath)

	// Check if the stream was opened successfully
	if result <= 0 || c.stream.State <= 0 {
		return fmt.Errorf("failed to connect to NTRIP server: %s", c.stream.Msg)
	}

	c.connected = true
	return nil
}

// Disconnect disconnects from the NTRIP server
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	c.stream.StreamClose()
	c.connected = false
	return nil
}

// Read reads data from the NTRIP server
func (c *Client) Read(p []byte) (int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	n := c.stream.StreamRead(p, len(p))
	if n <= 0 {
		return 0, io.EOF
	}

	return n, nil
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.connected
}

// GetStream returns the underlying stream
func (c *Client) GetStream() *gnssgo.Stream {
	return &c.stream
}
