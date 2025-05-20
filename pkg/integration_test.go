package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/caster"
	"github.com/bramburn/gnssgo/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestServerCasterIntegration tests that the server and caster work together
func TestServerCasterIntegration(t *testing.T) {
	// Create a logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a source service
	svc := caster.NewInMemorySourceService()
	svc.Sourcetable = caster.Sourcetable{
		Mounts: []caster.StreamEntry{
			{
				Name:       "TEST",
				Identifier: "TEST",
				Format:     "RTCM 3.3",
			},
		},
	}

	// Create a caster
	caster := caster.NewCaster(":2102", svc, logger)

	// Start the caster in a goroutine
	go func() {
		if err := caster.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Caster error: %v", err)
		}
	}()

	// Wait for the caster to start
	time.Sleep(100 * time.Millisecond)

	// Create a mock data source
	dataSource := &MockDataSource{
		dataChan: make(chan []byte, 10),
		data:     []byte("test data"),
	}

	// Create a server
	server := server.NewServer("localhost", "2102", "admin", "password", "TEST", logger)

	// Set the data source
	server.SetDataSource(dataSource)

	// Start the server
	err := server.Start()
	assert.NoError(t, err)
	defer server.Stop()

	// Wait for the server to connect to the caster
	time.Sleep(100 * time.Millisecond)

	// Create a client to subscribe to the caster
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "http://localhost:2102/TEST", nil)
	assert.NoError(t, err)
	req.SetBasicAuth("user", "password")

	// Send the request
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Read data from the response
	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "test data", string(buffer[:n]))

	// Shutdown the caster
	caster.Shutdown(ctx)
}

// MockDataSource is a mock data source for testing
type MockDataSource struct {
	dataChan chan []byte
	running  bool
	data     []byte
}

// Start starts the data source
func (ds *MockDataSource) Start() error {
	if ds.running {
		return nil
	}

	// Start generating data in a goroutine
	go func() {
		// Send the data to the channel
		ds.dataChan <- ds.data
	}()

	ds.running = true
	return nil
}

// Stop stops the data source
func (ds *MockDataSource) Stop() error {
	if !ds.running {
		return nil
	}

	// Close the data channel
	close(ds.dataChan)

	ds.running = false
	return nil
}

// Data returns the data channel
func (ds *MockDataSource) Data() <-chan []byte {
	return ds.dataChan
}
