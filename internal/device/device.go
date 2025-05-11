package device

import (
	"time"

	"github.com/bramburn/go_ntrip/internal/parser"
)

// Protocol constants
const (
	ProtocolNMEA = "NMEA-0183"
	ProtocolRTCM = "RTCM3.3"
	ProtocolUBX  = "UBX"
)

// GNSSDevice defines the interface for GNSS device operations
type GNSSDevice interface {
	// Connect establishes a connection to the device
	Connect(portName string, baudRate int) error

	// Disconnect closes the connection to the device
	Disconnect() error

	// IsConnected returns whether the device is connected
	IsConnected() bool

	// VerifyConnection checks if the device is sending valid GNSS data
	VerifyConnection(timeout time.Duration) bool

	// ReadRaw reads raw data from the device
	ReadRaw(buffer []byte) (int, error)

	// WriteCommand sends a command to the device
	WriteCommand(command string) error

	// ChangeBaudRate changes the baud rate of the connection
	ChangeBaudRate(baudRate int) error

	// GetAvailablePorts returns a list of available serial ports
	GetAvailablePorts() ([]string, error)

	// GetPortDetails returns detailed information about available ports
	GetPortDetails() ([]PortDetail, error)
}

// PortDetail represents details about a serial port
type PortDetail struct {
	Name    string
	IsUSB   bool
	VID     uint16
	PID     uint16
	Product string
}

// DataHandler defines the interface for handling data from the device
type DataHandler interface {
	// HandleNMEA handles NMEA sentences
	HandleNMEA(sentence parser.NMEASentence)

	// HandleRTCM handles RTCM messages
	HandleRTCM(message parser.RTCMMessage)

	// HandleUBX handles UBX messages
	HandleUBX(message parser.UBXMessage)
}

// MonitorConfig holds configuration for monitoring
type MonitorConfig struct {
	Protocol     string        // Protocol to monitor (NMEA, RTCM, UBX)
	BufferSize   int           // Size of the read buffer
	PollInterval time.Duration // Interval between reads
	Handler      DataHandler   // Handler for processed data
}

// DefaultMonitorConfig returns a default monitoring configuration
func DefaultMonitorConfig(protocol string, handler DataHandler) MonitorConfig {
	bufferSize := 1024
	if protocol == ProtocolRTCM {
		bufferSize = 2048 // RTCM messages can be larger
	}

	return MonitorConfig{
		Protocol:     protocol,
		BufferSize:   bufferSize,
		PollInterval: 100 * time.Millisecond,
		Handler:      handler,
	}
}
