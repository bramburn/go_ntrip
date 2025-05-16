package rtk

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/bramburn/go_ntrip/internal/position"
	"github.com/go-gnss/rtcm/rtcm3"
	"github.com/adrianmo/go-nmea"
)

// Solution status constants
const (
	StatusNone   = 0 // No solution
	StatusSingle = 1 // Single solution
	StatusDGPS   = 2 // DGPS solution
	StatusFloat  = 5 // Float RTK solution
	StatusFix    = 4 // Fixed RTK solution
)

// RTKSolution represents a solution from RTK processing
type RTKSolution struct {
	Status    int       // Solution status (None, Single, DGPS, Float, Fix)
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	Time      time.Time // Time of solution
	NumSats   int       // Number of satellites used
	HDOP      float64   // Horizontal dilution of precision
}

// Processor handles RTK processing of GNSS data
type Processor struct {
	mutex         sync.Mutex
	rtcmData      []byte
	solutions     []RTKSolution
	lastSolution  *RTKSolution
	solutionChan  chan RTKSolution
	processingRun bool
	mode          string // "static" or "kinematic"
	// RTK state data
	basePosition  *position.Position // Base station position (for static mode)
	rtcmMessages  map[uint16]rtcm3.Message // Store latest RTCM messages by type
	observations  []rtcm3.Message1004 // Store GPS observations
	ephemeris     map[int]rtcm3.Message1019 // Store GPS ephemeris by satellite ID
}

// NewProcessor creates a new RTK processor with default kinematic mode
func NewProcessor() *Processor {
	return NewProcessorWithMode("kinematic")
}

// NewProcessorWithMode creates a new RTK processor with the specified mode
func NewProcessorWithMode(mode string) *Processor {
	// Validate mode
	if mode != "static" && mode != "kinematic" {
		mode = "kinematic" // Default to kinematic if invalid mode
	}

	return &Processor{
		rtcmData:     make([]byte, 0),
		solutions:    make([]RTKSolution, 0),
		solutionChan: make(chan RTKSolution, 10),
		mode:         mode,
		rtcmMessages: make(map[uint16]rtcm3.Message),
		ephemeris:    make(map[int]rtcm3.Message1019),
	}
}

// ProcessRTCM processes RTCM data
func (p *Processor) ProcessRTCM(data []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Append new RTCM data
	p.rtcmData = append(p.rtcmData, data...)

	// Process RTCM data using the rtcm library
	if len(p.rtcmData) > 0 {
		// Try to parse RTCM messages
		messages, err := p.parseRTCMData(p.rtcmData)
		if err == nil && len(messages) > 0 {
			// Process the parsed messages
			p.processRTCMMessages(messages)

			// Compute RTK solution
			solution, err := p.computeRTKSolution()
			if err == nil {
				// Store the solution
				p.solutions = append(p.solutions, solution)
				p.lastSolution = &solution

				// Send the solution to the channel
				select {
				case p.solutionChan <- solution:
					// Solution sent successfully
				default:
					// Channel is full, discard the solution
				}
			}
		}

		// Clear the buffer after processing
		p.rtcmData = make([]byte, 0)
	}
}

// parseRTCMData parses RTCM data into messages
func (p *Processor) parseRTCMData(data []byte) ([]rtcm3.Message, error) {
	var messages []rtcm3.Message

	// Create a frame parser
	parser := rtcm3.NewParser()

	// Add data to the parser
	parser.Write(data)

	// Parse frames
	for {
		frame, err := parser.NextFrame()
		if err != nil {
			break // No more complete frames
		}

		// Parse the message from the frame
		msg, err := rtcm3.DeserializeMessage(frame.Data)
		if err != nil {
			continue // Skip invalid messages
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// processRTCMMessages processes RTCM messages and updates internal state
func (p *Processor) processRTCMMessages(messages []rtcm3.Message) {
	for _, msg := range messages {
		// Store the message by type
		p.rtcmMessages[msg.Number()] = msg

		// Process specific message types
		switch msg.Number() {
		case 1004: // GPS L1/L2 observations
			if obsMsg, ok := msg.(rtcm3.Message1004); ok {
				p.observations = append(p.observations, obsMsg)
				// Keep only the last 10 observation messages
				if len(p.observations) > 10 {
					p.observations = p.observations[len(p.observations)-10:]
				}
			}

		case 1019: // GPS ephemeris
			if ephMsg, ok := msg.(rtcm3.Message1019); ok {
				p.ephemeris[int(ephMsg.SatelliteID)] = ephMsg
			}

		case 1005, 1006: // Station coordinates
			// If in static mode, use this as the base position
			if p.mode == "static" {
				if stationMsg, ok := msg.(rtcm3.Message1005); ok {
					// Convert ECEF to lat/lon/alt
					lat, lon, alt := ecefToLatLonAlt(stationMsg.X, stationMsg.Y, stationMsg.Z)

					// Update base position
					p.basePosition = &position.Position{
						Latitude:    lat,
						Longitude:   lon,
						Altitude:    alt,
						FixQuality:  StatusFix,
						Timestamp:   time.Now().UTC(),
						Description: "RTCM Base Station Position",
					}
				}
			}
		}
	}
}

// computeRTKSolution computes an RTK solution based on the current state
func (p *Processor) computeRTKSolution() (RTKSolution, error) {
	// In a real implementation, this would use the collected RTCM data
	// to compute a precise RTK solution

	// For static mode, if we have a base position, use it
	if p.mode == "static" && p.basePosition != nil {
		return RTKSolution{
			Status:    StatusFix,
			Latitude:  p.basePosition.Latitude,
			Longitude: p.basePosition.Longitude,
			Altitude:  p.basePosition.Altitude,
			Time:      time.Now().UTC(),
			NumSats:   len(p.ephemeris), // Use number of satellites with ephemeris
			HDOP:      0.8, // Placeholder
		}, nil
	}

	// For kinematic mode or if no base position is available
	// Use the observations and ephemeris to compute a solution
	if len(p.observations) > 0 {
		// Get the latest observation
		latestObs := p.observations[len(p.observations)-1]

		// Count satellites with valid observations
		numSats := len(latestObs.Satellites)

		// In a real implementation, we would compute a solution using
		// the observations and ephemeris data

		// For now, return a simulated solution based on the data we have
		return RTKSolution{
			Status:    computeFixStatus(numSats),
			Latitude:  51.5074, // Placeholder
			Longitude: -0.1278, // Placeholder
			Altitude:  45.0,    // Placeholder
			Time:      time.Now().UTC(),
			NumSats:   numSats,
			HDOP:      computeHDOP(numSats),
		}, nil
	}

	// If we don't have enough data, return an error
	return RTKSolution{}, fmt.Errorf("not enough data for RTK solution")
}

// computeFixStatus determines the fix status based on available data
func computeFixStatus(numSats int) int {
	if numSats >= 5 {
		return StatusFix
	} else if numSats >= 4 {
		return StatusFloat
	} else if numSats >= 3 {
		return StatusDGPS
	} else if numSats >= 2 {
		return StatusSingle
	}
	return StatusNone
}

// computeHDOP calculates HDOP based on number of satellites
func computeHDOP(numSats int) float64 {
	if numSats <= 0 {
		return 99.9
	}
	// Simple approximation - in reality this depends on satellite geometry
	return 5.0 / float64(numSats)
}

// ecefToLatLonAlt converts ECEF coordinates to latitude, longitude, altitude
func ecefToLatLonAlt(x, y, z float64) (lat, lon, alt float64) {
	// WGS84 ellipsoid parameters
	a := 6378137.0 // semi-major axis
	f := 1.0 / 298.257223563 // flattening
	b := a * (1.0 - f) // semi-minor axis
	e2 := f * (2.0 - f) // eccentricity squared

	// Calculate longitude
	lon = math.Atan2(y, x) * 180.0 / math.Pi

	// Calculate latitude and altitude iteratively
	p := math.Sqrt(x*x + y*y)
	lat = math.Atan2(z, p * (1.0 - e2))

	// Iterative calculation for better accuracy
	for i := 0; i < 5; i++ {
		sinLat := math.Sin(lat)
		N := a / math.Sqrt(1.0 - e2*sinLat*sinLat)
		alt = p/math.Cos(lat) - N
		lat = math.Atan2(z, p*(1.0-e2*N/(N+alt)))
	}

	// Convert latitude to degrees
	lat = lat * 180.0 / math.Pi

	return lat, lon, alt
}

// GetSolutionChannel returns the channel for receiving solutions
func (p *Processor) GetSolutionChannel() <-chan RTKSolution {
	return p.solutionChan
}

// GetLastSolution returns the last computed solution
func (p *Processor) GetLastSolution() *RTKSolution {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.lastSolution == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	solution := *p.lastSolution
	return &solution
}

// ToPosition converts an RTK solution to a Position
func (s *RTKSolution) ToPosition() *position.Position {
	return &position.Position{
		Latitude:    s.Latitude,
		Longitude:   s.Longitude,
		Altitude:    s.Altitude,
		FixQuality:  s.Status,
		Satellites:  s.NumSats,
		HDOP:        s.HDOP,
		Timestamp:   s.Time,
		Description: fmt.Sprintf("RTK Solution: %s", position.GetFixQualityDescription(s.Status)),
	}
}

// StartProcessing starts continuous RTK processing
func (p *Processor) StartProcessing() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.processingRun {
		return
	}

	p.processingRun = true

	// In a real implementation, this would start a goroutine that continuously
	// processes GNSS data and computes RTK solutions
}

// StopProcessing stops RTK processing
func (p *Processor) StopProcessing() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.processingRun {
		return
	}

	p.processingRun = false

	// In a real implementation, this would stop the processing goroutine
}
