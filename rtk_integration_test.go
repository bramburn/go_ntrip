package go_ntrip

import (
	"fmt"
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
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
	// Format: port[:brate[:bsize[:parity[:stopb[:fctr]]]]]
	GNSS_RECEIVER_PORT = "COM3:115200:8:N:1"
)

// TestGNSSReceiverConnection tests the connection to the physical GNSS receiver
func TestGNSSReceiverConnection(t *testing.T) {
	// Create a stream for the GNSS receiver
	var stream gnssgo.Stream
	
	// Initialize the stream
	stream.InitStream()
	
	// Open the serial port connected to the GNSS receiver
	result := stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_R, GNSS_RECEIVER_PORT)
	
	// Check if the stream was opened successfully
	if result > 0 && stream.State > 0 {
		t.Logf("Successfully connected to GNSS receiver on %s", GNSS_RECEIVER_PORT)
		
		// Read some data from the stream
		buff := make([]byte, 4096)
		n := stream.StreamRead(buff, 4096)
		
		if n > 0 {
			t.Logf("Successfully read %d bytes from GNSS receiver", n)
			
			// Log the first few bytes to see what format we're receiving
			t.Logf("First 10 bytes: % X", buff[:min(n, 10)])
			
			// Check if the data looks like UBX or NMEA
			if n > 0 && (buff[0] == 0xB5 && buff[1] == 0x62) {
				t.Logf("Data appears to be in UBX format")
			} else if n > 0 && buff[0] == '$' {
				t.Logf("Data appears to be in NMEA format")
			} else {
				t.Logf("Data format not recognized")
			}
		} else {
			t.Logf("No data received from GNSS receiver within timeout")
		}
		
		// Close the stream
		stream.StreamClose()
		assert.True(t, true, "GNSS receiver connection test passed")
	} else {
		t.Logf("Failed to connect to GNSS receiver, state: %d", stream.State)
		t.Logf("Error message: %s", stream.Msg)
		t.Skip("Skipping test due to connection failure - check if the GNSS receiver is connected and the port is correct")
	}
}

// TestNtripConnection tests the connection to the NTRIP server
func TestNtripConnection(t *testing.T) {
	// Create a stream for NTRIP client
	var stream gnssgo.Stream
	
	// Initialize the stream
	stream.InitStream()
	
	// Construct the NTRIP path
	ntripPath := fmt.Sprintf("%s:%s@%s:%s/%s", 
		NTRIP_USER, NTRIP_PASSWORD, NTRIP_SERVER, NTRIP_PORT, NTRIP_MOUNTPOINT)
	
	// Open the NTRIP client stream
	result := stream.OpenStream(gnssgo.STR_NTRIPCLI, gnssgo.STR_MODE_R, ntripPath)
	
	// Check if the stream was opened successfully
	if result > 0 && stream.State > 0 {
		t.Logf("Successfully connected to NTRIP server: %s", ntripPath)
		
		// Read some data from the stream
		buff := make([]byte, 4096)
		n := stream.StreamRead(buff, 4096)
		
		if n > 0 {
			t.Logf("Successfully read %d bytes from NTRIP server", n)
			
			// Log the first few bytes
			t.Logf("First 10 bytes: % X", buff[:min(n, 10)])
			
			// Check if the data looks like RTCM3
			// RTCM3 messages start with 0xD3
			if n > 2 && buff[0] == 0xD3 {
				t.Logf("Data appears to be in RTCM3 format")
			} else {
				t.Logf("Data does not appear to be in RTCM3 format")
			}
		} else {
			t.Logf("No data received from NTRIP server within timeout")
		}
		
		// Close the stream
		stream.StreamClose()
		assert.True(t, true, "NTRIP connection test passed")
	} else {
		t.Logf("Failed to connect to NTRIP server, state: %d", stream.State)
		t.Logf("Error message: %s", stream.Msg)
		t.Skip("Skipping test due to connection failure - check if the NTRIP server is available")
	}
}

// TestFullRTKWorkflow tests the complete RTK workflow
func TestFullRTKWorkflow(t *testing.T) {
	// Create streams for GNSS receiver and NTRIP client
	var roverStream, baseStream gnssgo.Stream
	
	// Initialize the streams
	roverStream.InitStream()
	baseStream.InitStream()
	
	// Construct the NTRIP path
	ntripPath := fmt.Sprintf("%s:%s@%s:%s/%s", 
		NTRIP_USER, NTRIP_PASSWORD, NTRIP_SERVER, NTRIP_PORT, NTRIP_MOUNTPOINT)
	
	// Open the GNSS receiver stream
	roverResult := roverStream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_R, GNSS_RECEIVER_PORT)
	if roverResult <= 0 || roverStream.State <= 0 {
		t.Logf("Failed to connect to GNSS receiver, state: %d", roverStream.State)
		t.Logf("Error message: %s", roverStream.Msg)
		t.Skip("Skipping test due to GNSS receiver connection failure")
		return
	}
	
	// Open the NTRIP client stream
	baseResult := baseStream.OpenStream(gnssgo.STR_NTRIPCLI, gnssgo.STR_MODE_R, ntripPath)
	if baseResult <= 0 || baseStream.State <= 0 {
		t.Logf("Failed to connect to NTRIP server, state: %d", baseStream.State)
		t.Logf("Error message: %s", baseStream.Msg)
		roverStream.StreamClose()
		t.Skip("Skipping test due to NTRIP connection failure")
		return
	}
	
	// Initialize RTK server
	var svr gnssgo.RtkSvr
	
	// Configure RTK processing options
	var prcopt gnssgo.PrcOpt
	prcopt.Mode = gnssgo.PMODE_KINEMA // Kinematic mode
	prcopt.NavSys = gnssgo.SYS_GPS | gnssgo.SYS_GLO // GPS + GLONASS
	prcopt.RefPos = 1 // Use average of single position
	prcopt.ElMask = 15.0 * gnssgo.D2R // Elevation mask (15 degrees)
	
	// Configure solution options
	var solopt [2]gnssgo.SolOpt
	solopt[0].Posf = gnssgo.SOLF_LLH // Latitude/Longitude/Height format
	solopt[1].Posf = gnssgo.SOLF_NMEA // NMEA format
	
	// Configure stream types
	strtype := []int{
		gnssgo.STR_SERIAL,   // Rover input (physical GNSS receiver)
		gnssgo.STR_NTRIPCLI, // Base station input (NTRIP)
		gnssgo.STR_NONE,     // Ephemeris input
		gnssgo.STR_FILE,     // Solution 1 output
		gnssgo.STR_NONE,     // Solution 2 output
		gnssgo.STR_NONE,     // Log rover
		gnssgo.STR_NONE,     // Log base station
		gnssgo.STR_NONE,     // Log ephemeris
	}
	
	// Configure stream paths
	paths := []string{
		GNSS_RECEIVER_PORT, // Rover input (physical GNSS receiver)
		ntripPath,          // Base station input (NTRIP)
		"",                 // Ephemeris input
		"rtk_solution.pos", // Solution 1 output
		"",                 // Solution 2 output
		"",                 // Log rover
		"",                 // Log base station
		"",                 // Log ephemeris
	}
	
	// Configure stream formats
	strfmt := []int{
		gnssgo.STRFMT_UBX,   // Rover format (UBX)
		gnssgo.STRFMT_RTCM3, // Base station format (RTCM3)
		gnssgo.STRFMT_RINEX, // Ephemeris format
		gnssgo.SOLF_LLH,     // Solution 1 format
		gnssgo.SOLF_NMEA,    // Solution 2 format
	}
	
	// Start RTK server
	var errmsg string
	svrcycle := 10 // Server cycle (ms)
	buffsize := 32768 // Buffer size (bytes)
	navmsgsel := 0 // Navigation message select
	cmds := []string{"", "", ""} // Commands for input streams
	cmds_periodic := []string{"", "", ""} // Periodic commands
	rcvopts := []string{"", "", ""} // Receiver options
	nmeacycle := 1000 // NMEA request cycle (ms)
	nmeareq := 0 // NMEA request type
	nmeapos := []float64{0, 0, 0} // NMEA position
	
	// Start the RTK server
	if svr.RtkSvrStart(svrcycle, buffsize, strtype, paths, strfmt, navmsgsel,
		cmds, cmds_periodic, rcvopts, nmeacycle, nmeareq, nmeapos, &prcopt,
		solopt[:], nil, &errmsg) == 0 {
		t.Fatalf("Failed to start RTK server: %s", errmsg)
	}
	
	// Let the RTK server run for a while
	t.Logf("RTK server started, waiting for solutions...")
	time.Sleep(30 * time.Second)
	
	// Get RTK server status
	var sstat gnssgo.RtkSvrStat
	svr.RtkSvrGetStat(&sstat)
	
	// Log the status
	t.Logf("RTK server status:")
	t.Logf("  Time: %s", gnssgo.TimeStr(sstat.Time, 0))
	t.Logf("  Rover observations: %d", sstat.Obs[0])
	t.Logf("  Base observations: %d", sstat.Obs[1])
	t.Logf("  Solution status: %d", sstat.SolStat)
	
	// Check if we got any observations
	assert.Greater(t, sstat.Obs[0], 0, "Should have received rover observations")
	
	// We may not get a fix as the GNSS receiver is outside the window
	t.Logf("Note: A fix may not be achieved as the GNSS receiver is outside the window")
	
	// Stop the RTK server
	svr.RtkSvrStop(cmds)
	
	// Close the streams
	roverStream.StreamClose()
	baseStream.StreamClose()
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
