package ntrip_test

import (
	"testing"
	"time"

	"github.com/bramburn/go_ntrip/pkg/ntrip"
	"github.com/stretchr/testify/assert"
)

const (
	// NTRIP server configuration
	NTRIP_SERVER     = "192.168.0.64"
	NTRIP_PORT       = "2101"
	NTRIP_USER       = "reach"
	NTRIP_PASSWORD   = "emlidreach"
	NTRIP_MOUNTPOINT = "REACH"
	
	// GNSS receiver configuration - update with your actual COM port
	GNSS_RECEIVER_PORT = "COM3:115200:8:N:1"
)

// TestNtripConnection tests the connection to the NTRIP server
func TestNtripConnection(t *testing.T) {
	// Create a new NTRIP client
	client, err := ntrip.NewClient(NTRIP_SERVER, NTRIP_PORT, NTRIP_USER, NTRIP_PASSWORD, NTRIP_MOUNTPOINT)
	assert.NoError(t, err, "Should create NTRIP client without error")
	
	// Connect to the NTRIP server
	err = client.Connect()
	if err != nil {
		t.Logf("Failed to connect to NTRIP server: %v", err)
		t.Skip("Skipping test due to connection failure - check if the NTRIP server is available")
		return
	}
	defer client.Disconnect()
	
	// Read some data from the NTRIP server
	buffer := make([]byte, 4096)
	n, err := client.Read(buffer)
	
	if err != nil {
		t.Logf("Error reading from NTRIP server: %v", err)
	} else {
		t.Logf("Successfully read %d bytes from NTRIP server", n)
		
		// Log the first few bytes
		if n > 0 {
			t.Logf("First 10 bytes: % X", buffer[:min(n, 10)])
			
			// Check if the data looks like RTCM3
			// RTCM3 messages start with 0xD3
			if n > 2 && buffer[0] == 0xD3 {
				t.Logf("Data appears to be in RTCM3 format")
			} else {
				t.Logf("Data does not appear to be in RTCM3 format")
			}
		}
	}
	
	assert.True(t, client.IsConnected(), "Client should be connected")
}

// TestGNSSReceiverConnection tests the connection to the physical GNSS receiver
func TestGNSSReceiverConnection(t *testing.T) {
	// Create a new GNSS receiver
	receiver, err := ntrip.NewGNSSReceiver(GNSS_RECEIVER_PORT)
	if err != nil {
		t.Logf("Failed to connect to GNSS receiver: %v", err)
		t.Skip("Skipping test due to connection failure - check if the GNSS receiver is connected and the port is correct")
		return
	}
	defer receiver.Close()
	
	// Read some data from the GNSS receiver
	buffer := make([]byte, 4096)
	n, err := receiver.Read(buffer)
	
	if err != nil {
		t.Logf("Error reading from GNSS receiver: %v", err)
	} else {
		t.Logf("Successfully read %d bytes from GNSS receiver", n)
		
		// Log the first few bytes
		if n > 0 {
			t.Logf("First 10 bytes: % X", buffer[:min(n, 10)])
			
			// Check if the data looks like UBX or NMEA
			if n > 1 && buffer[0] == 0xB5 && buffer[1] == 0x62 {
				t.Logf("Data appears to be in UBX format")
			} else if n > 0 && buffer[0] == '$' {
				t.Logf("Data appears to be in NMEA format")
			} else {
				t.Logf("Data format not recognized")
			}
		}
	}
	
	assert.True(t, receiver.IsOpen(), "Receiver should be open")
}

// TestFullRTKWorkflow tests the complete RTK workflow
func TestFullRTKWorkflow(t *testing.T) {
	// Create a new GNSS receiver
	receiver, err := ntrip.NewGNSSReceiver(GNSS_RECEIVER_PORT)
	if err != nil {
		t.Logf("Failed to connect to GNSS receiver: %v", err)
		t.Skip("Skipping test due to connection failure - check if the GNSS receiver is connected and the port is correct")
		return
	}
	defer receiver.Close()
	
	// Create a new NTRIP client
	client, err := ntrip.NewClient(NTRIP_SERVER, NTRIP_PORT, NTRIP_USER, NTRIP_PASSWORD, NTRIP_MOUNTPOINT)
	assert.NoError(t, err, "Should create NTRIP client without error")
	
	// Connect to the NTRIP server
	err = client.Connect()
	if err != nil {
		t.Logf("Failed to connect to NTRIP server: %v", err)
		t.Skip("Skipping test due to connection failure - check if the NTRIP server is available")
		return
	}
	defer client.Disconnect()
	
	// Create a new RTK processor
	processor, err := ntrip.NewRTKProcessor(receiver, client)
	assert.NoError(t, err, "Should create RTK processor without error")
	
	// Start the RTK processing
	err = processor.Start()
	assert.NoError(t, err, "Should start RTK processing without error")
	
	// Let the RTK processor run for a while
	t.Logf("RTK processing started, waiting for solutions...")
	time.Sleep(30 * time.Second)
	
	// Get the RTK statistics
	stats := processor.GetStats()
	
	// Log the statistics
	t.Logf("RTK processing statistics:")
	t.Logf("  Rover observations: %d", stats.RoverObs)
	t.Logf("  Base observations: %d", stats.BaseObs)
	t.Logf("  Solutions: %d", stats.Solutions)
	t.Logf("  Fix ratio: %.2f%%", stats.FixRatio*100)
	
	// Check if we got any observations
	assert.Greater(t, stats.RoverObs, 0, "Should have received rover observations")
	
	// We may not get a fix as the GNSS receiver is outside the window
	t.Logf("Note: A fix may not be achieved as the GNSS receiver is outside the window")
	
	// Stop the RTK processing
	err = processor.Stop()
	assert.NoError(t, err, "Should stop RTK processing without error")
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
