package rtk

import (
	"fmt"
	"sync"
	"time"

	"github.com/bramburn/go_ntrip/internal/position"
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
}

// NewProcessor creates a new RTK processor
func NewProcessor() *Processor {
	return &Processor{
		rtcmData:     make([]byte, 0),
		solutions:    make([]RTKSolution, 0),
		solutionChan: make(chan RTKSolution, 10),
	}
}

// ProcessRTCM processes RTCM data
func (p *Processor) ProcessRTCM(data []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Append new RTCM data
	p.rtcmData = append(p.rtcmData, data...)

	// In a real implementation, we would parse the RTCM data and update the RTK solution
	// For now, we'll simulate RTK processing with a simple solution
	if len(p.rtcmData) > 1024 {
		// Simulate processing - in a real implementation, this would use the gnssgo library
		p.simulateRTKSolution()

		// Clear the buffer after processing
		p.rtcmData = make([]byte, 0)
	}
}

// simulateRTKSolution simulates an RTK solution for demonstration purposes
// In a real implementation, this would use the gnssgo library to compute a solution
func (p *Processor) simulateRTKSolution() {
	// Create a simulated solution
	solution := RTKSolution{
		Status:    StatusFix, // Simulate a fixed solution
		Latitude:  51.5074,   // Example latitude (London)
		Longitude: -0.1278,   // Example longitude (London)
		Altitude:  45.0,      // Example altitude
		Time:      time.Now().UTC(),
		NumSats:   12,  // Example satellite count
		HDOP:      0.8, // Example HDOP
	}

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
