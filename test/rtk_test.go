package test

import (
	"testing"
	"time"

	"github.com/bramburn/go_ntrip/internal/position"
	"github.com/bramburn/go_ntrip/internal/rtk"
)

func TestRTKIntegration(t *testing.T) {
	// Create RTK processor with default kinematic mode
	processor := rtk.NewProcessor()

	// Create position averager
	averager := position.NewPositionAverager(4)

	// Process some RTCM data
	rtcmData := make([]byte, 2000)
	for i := range rtcmData {
		rtcmData[i] = byte(i % 256)
	}

	// Process data multiple times to generate multiple solutions
	for i := 0; i < 5; i++ {
		processor.ProcessRTCM(rtcmData)

		// Wait for solution to be generated
		time.Sleep(10 * time.Millisecond)

		// Get last solution
		solution := processor.GetLastSolution()
		if solution == nil {
			t.Fatal("Expected non-nil solution")
		}

		// Convert to position
		pos := solution.ToPosition()

		// Create sample
		sample := position.PositionSample{
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Altitude:   pos.Altitude,
			FixQuality: pos.FixQuality,
			Timestamp:  pos.Timestamp,
		}

		// Add sample to averager
		averager.AddSample(sample)
	}

	// Check that samples were added
	if averager.GetSampleCount() != 5 {
		t.Errorf("Expected 5 samples, got %d", averager.GetSampleCount())
	}

	// Get averaged position
	pos, stats, err := averager.GetAveragedPosition()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check position
	if pos == nil {
		t.Fatal("Expected non-nil position")
	}

	// Check stats
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats.SampleCount != 5 {
		t.Errorf("Expected sample count 5, got %d", stats.SampleCount)
	}
}

func TestRTKProcessorSolutionChannel(t *testing.T) {
	// Create RTK processor with default kinematic mode
	processor := rtk.NewProcessor()

	// Start processing
	processor.StartProcessing()
	defer processor.StopProcessing()

	// Get solution channel
	solutionChan := processor.GetSolutionChannel()

	// Process some RTCM data
	rtcmData := make([]byte, 2000)
	processor.ProcessRTCM(rtcmData)

	// Wait for solution
	select {
	case solution := <-solutionChan:
		// Check solution
		if solution.Status != rtk.StatusFix {
			t.Errorf("Expected status %d, got %d", rtk.StatusFix, solution.Status)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for solution")
	}
}

func TestRTKStaticMode(t *testing.T) {
	// Create RTK processor with static mode
	processor := rtk.NewProcessorWithMode("static")

	// Verify mode is set correctly
	if processor.GetMode() != "static" {
		t.Errorf("Expected mode 'static', got '%s'", processor.GetMode())
	}

	// Process some RTCM data
	rtcmData := make([]byte, 2000)
	for i := range rtcmData {
		rtcmData[i] = byte(i % 256)
	}

	// Process data
	processor.ProcessRTCM(rtcmData)

	// Wait for solution to be generated
	time.Sleep(10 * time.Millisecond)

	// Get last solution
	solution := processor.GetLastSolution()
	if solution == nil {
		t.Fatal("Expected non-nil solution")
	}

	// Convert to position
	pos := solution.ToPosition()

	// Verify position has appropriate fix quality for static mode
	if pos.FixQuality < rtk.StatusDGPS {
		t.Errorf("Expected fix quality >= %d, got %d", rtk.StatusDGPS, pos.FixQuality)
	}
}