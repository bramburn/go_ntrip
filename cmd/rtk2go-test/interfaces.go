package main

import (
	"time"
)

// GNSSDevice defines the interface for GNSS device operations
type GNSSDevice interface {
	// Connect establishes a connection to the device
	Connect() error

	// Disconnect closes the connection to the device
	Disconnect() error

	// IsConnected returns whether the device is connected
	IsConnected() bool

	// VerifyConnection checks if the device is sending valid GNSS data
	VerifyConnection(timeout time.Duration) bool

	// ReadRaw reads raw data from the device
	ReadRaw(buffer []byte) (int, error)

	// WriteRaw writes raw data to the device
	WriteRaw(data []byte) (int, error)
}

// NTRIPClient defines the interface for NTRIP client operations
type NTRIPClient interface {
	// Connect connects to the NTRIP server
	Connect() error

	// Disconnect disconnects from the NTRIP server
	Disconnect() error

	// IsConnected returns whether the client is connected
	IsConnected() bool

	// Read reads data from the NTRIP server
	Read(buffer []byte) (int, error)

	// Write writes data to the NTRIP server
	Write(data []byte) (int, error)
}

// RTKProcessor defines the interface for RTK processing operations
type RTKProcessor interface {
	// Start starts the RTK processing
	Start() error

	// Stop stops the RTK processing
	Stop() error

	// GetSolution returns the current RTK solution
	GetSolution() RTKSolution

	// GetStats returns the current RTK statistics
	GetStats() RTKStats
}

// NMEAParser defines the interface for NMEA parsing operations
type NMEAParser interface {
	// Parse parses an NMEA sentence
	Parse(sentence string) (NMEASentence, error)

	// ParseGGA parses a GGA sentence
	ParseGGA(sentence string) (GGAData, error)
}

// NMEASentence represents a parsed NMEA sentence
type NMEASentence struct {
	Raw      string
	Type     string
	Fields   []string
	Valid    bool
	Checksum string
}

// GGAData represents parsed data from a GGA sentence
type GGAData struct {
	Time      string
	Latitude  float64
	LatDir    string
	Longitude float64
	LonDir    string
	Quality   int
	NumSats   int
	HDOP      float64
	Altitude  float64
	AltUnit   string
	GeoidSep  float64
	GeoidUnit string
	DGPSAge   float64
	DGPSStaID string
}

// RTKStats represents RTK processing statistics
type RTKStats struct {
	Solutions int     // Number of solutions processed
	FixRatio  float64 // Ratio of fixed solutions to total solutions
}

// Logger defines a simple logging interface
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatalf(format string, v ...interface{})
}
