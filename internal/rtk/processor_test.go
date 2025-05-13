package rtk

import (
	"testing"
	"time"
)

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor()
	if processor == nil {
		t.Fatal("NewProcessor returned nil")
	}

	if processor.rtcmData == nil {
		t.Error("rtcmData should be initialized")
	}

	if processor.solutions == nil {
		t.Error("solutions should be initialized")
	}

	if processor.solutionChan == nil {
		t.Error("solutionChan should be initialized")
	}
}

func TestProcessRTCM(t *testing.T) {
	processor := NewProcessor()

	// Test with small data (less than threshold)
	smallData := []byte("test data")
	processor.ProcessRTCM(smallData)

	if len(processor.rtcmData) != len(smallData) {
		t.Errorf("Expected rtcmData length %d, got %d", len(smallData), len(processor.rtcmData))
	}

	// Test with large data (more than threshold)
	largeData := make([]byte, 2000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	processor.ProcessRTCM(largeData)

	// After processing large data, the buffer should be cleared
	if len(processor.rtcmData) != 0 {
		t.Errorf("Expected rtcmData to be cleared, got length %d", len(processor.rtcmData))
	}

	// Should have generated a solution
	if len(processor.solutions) != 1 {
		t.Errorf("Expected 1 solution, got %d", len(processor.solutions))
	}
}

func TestGetSolutionChannel(t *testing.T) {
	processor := NewProcessor()

	channel := processor.GetSolutionChannel()
	if channel == nil {
		t.Error("GetSolutionChannel returned nil")
	}
}

func TestGetLastSolution(t *testing.T) {
	processor := NewProcessor()

	// Initially, there should be no solution
	solution := processor.GetLastSolution()
	if solution != nil {
		t.Error("Expected nil solution initially")
	}

	// Process some data to generate a solution
	largeData := make([]byte, 2000)
	processor.ProcessRTCM(largeData)

	// Now there should be a solution
	solution = processor.GetLastSolution()
	if solution == nil {
		t.Error("Expected non-nil solution after processing")
	}
}

func TestSolutionToPosition(t *testing.T) {
	solution := RTKSolution{
		Status:    StatusFix,
		Latitude:  51.5074,
		Longitude: -0.1278,
		Altitude:  45.0,
		Time:      time.Now().UTC(),
		NumSats:   12,
		HDOP:      0.8,
	}

	position := solution.ToPosition()

	if position == nil {
		t.Fatal("ToPosition returned nil")
	}

	if position.Latitude != solution.Latitude {
		t.Errorf("Expected latitude %f, got %f", solution.Latitude, position.Latitude)
	}

	if position.Longitude != solution.Longitude {
		t.Errorf("Expected longitude %f, got %f", solution.Longitude, position.Longitude)
	}

	if position.Altitude != solution.Altitude {
		t.Errorf("Expected altitude %f, got %f", solution.Altitude, position.Altitude)
	}

	if position.FixQuality != solution.Status {
		t.Errorf("Expected fix quality %d, got %d", solution.Status, position.FixQuality)
	}

	if position.Satellites != solution.NumSats {
		t.Errorf("Expected satellites %d, got %d", solution.NumSats, position.Satellites)
	}

	if position.HDOP != solution.HDOP {
		t.Errorf("Expected HDOP %f, got %f", solution.HDOP, position.HDOP)
	}
}

func TestStartStopProcessing(t *testing.T) {
	processor := NewProcessor()

	// Initially, processing should not be running
	if processor.processingRun {
		t.Error("Expected processingRun to be false initially")
	}

	// Start processing
	processor.StartProcessing()

	// Now processing should be running
	if !processor.processingRun {
		t.Error("Expected processingRun to be true after starting")
	}

	// Stop processing
	processor.StopProcessing()

	// Now processing should be stopped
	if processor.processingRun {
		t.Error("Expected processingRun to be false after stopping")
	}
}
