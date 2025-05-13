package position

import (
	"math"
	"testing"
	"time"
)

func TestNewPositionAverager(t *testing.T) {
	minFixQuality := 4
	averager := NewPositionAverager(minFixQuality)

	if averager == nil {
		t.Fatal("NewPositionAverager returned nil")
	}

	if averager.minFixQuality != minFixQuality {
		t.Errorf("Expected minFixQuality %d, got %d", minFixQuality, averager.minFixQuality)
	}

	if averager.samples == nil {
		t.Error("samples should be initialized")
	}

	if averager.fixQualityDist == nil {
		t.Error("fixQualityDist should be initialized")
	}
}

func TestAddSample(t *testing.T) {
	averager := NewPositionAverager(4)

	// Test with sample below minimum fix quality
	lowQualitySample := PositionSample{
		Latitude:   51.5074,
		Longitude:  -0.1278,
		Altitude:   45.0,
		FixQuality: 3,
		Timestamp:  time.Now().UTC(),
	}

	accepted := averager.AddSample(lowQualitySample)
	if accepted {
		t.Error("Expected sample with low fix quality to be rejected")
	}

	// The fix quality distribution should still be updated
	if averager.fixQualityDist[3] != 1 {
		t.Errorf("Expected fix quality distribution for quality 3 to be 1, got %d", averager.fixQualityDist[3])
	}

	// Test with sample at minimum fix quality
	goodQualitySample := PositionSample{
		Latitude:   51.5074,
		Longitude:  -0.1278,
		Altitude:   45.0,
		FixQuality: 4,
		Timestamp:  time.Now().UTC(),
	}

	accepted = averager.AddSample(goodQualitySample)
	if !accepted {
		t.Error("Expected sample with good fix quality to be accepted")
	}

	// The sample should be added
	if len(averager.samples) != 1 {
		t.Errorf("Expected 1 sample, got %d", len(averager.samples))
	}

	// The fix quality distribution should be updated
	if averager.fixQualityDist[4] != 1 {
		t.Errorf("Expected fix quality distribution for quality 4 to be 1, got %d", averager.fixQualityDist[4])
	}
}

func TestGetSampleCount(t *testing.T) {
	averager := NewPositionAverager(4)

	// Initially, there should be no samples
	if averager.GetSampleCount() != 0 {
		t.Errorf("Expected 0 samples initially, got %d", averager.GetSampleCount())
	}

	// Add a sample
	sample := PositionSample{
		Latitude:   51.5074,
		Longitude:  -0.1278,
		Altitude:   45.0,
		FixQuality: 4,
		Timestamp:  time.Now().UTC(),
	}

	averager.AddSample(sample)

	// Now there should be one sample
	if averager.GetSampleCount() != 1 {
		t.Errorf("Expected 1 sample after adding, got %d", averager.GetSampleCount())
	}
}

func TestGetAveragedPosition(t *testing.T) {
	averager := NewPositionAverager(4)

	// Test with no samples
	pos, stats, err := averager.GetAveragedPosition()
	if err == nil {
		t.Error("Expected error with no samples")
	}
	if pos != nil {
		t.Error("Expected nil position with no samples")
	}
	if stats != nil {
		t.Error("Expected nil stats with no samples")
	}

	// Add some samples
	now := time.Now().UTC()
	samples := []PositionSample{
		{
			Latitude:   51.5074,
			Longitude:  -0.1278,
			Altitude:   45.0,
			FixQuality: 4,
			Timestamp:  now,
		},
		{
			Latitude:   51.5076,
			Longitude:  -0.1276,
			Altitude:   46.0,
			FixQuality: 4,
			Timestamp:  now.Add(1 * time.Second),
		},
		{
			Latitude:   51.5078,
			Longitude:  -0.1274,
			Altitude:   47.0,
			FixQuality: 5,
			Timestamp:  now.Add(2 * time.Second),
		},
	}

	for _, sample := range samples {
		averager.AddSample(sample)
	}

	// Now get the averaged position
	pos, stats, err = averager.GetAveragedPosition()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pos == nil {
		t.Fatal("Expected non-nil position")
	}
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Check the averaged position
	expectedLat := (51.5074 + 51.5076 + 51.5078) / 3
	expectedLon := (-0.1278 + -0.1276 + -0.1274) / 3
	expectedAlt := (45.0 + 46.0 + 47.0) / 3

	if math.Abs(pos.Latitude-expectedLat) > 0.0001 {
		t.Errorf("Expected latitude %f, got %f", expectedLat, pos.Latitude)
	}

	if math.Abs(pos.Longitude-expectedLon) > 0.0001 {
		t.Errorf("Expected longitude %f, got %f", expectedLon, pos.Longitude)
	}

	if math.Abs(pos.Altitude-expectedAlt) > 0.0001 {
		t.Errorf("Expected altitude %f, got %f", expectedAlt, pos.Altitude)
	}

	// Check the stats
	if stats.SampleCount != 3 {
		t.Errorf("Expected sample count 3, got %d", stats.SampleCount)
	}

	if stats.Duration != 2.0 {
		t.Errorf("Expected duration 2.0, got %f", stats.Duration)
	}

	// Check fix quality distribution
	if stats.FixQualityDistribution[4] != 2 {
		t.Errorf("Expected 2 samples with fix quality 4, got %d", stats.FixQualityDistribution[4])
	}

	if stats.FixQualityDistribution[5] != 1 {
		t.Errorf("Expected 1 sample with fix quality 5, got %d", stats.FixQualityDistribution[5])
	}
}

func TestReset(t *testing.T) {
	averager := NewPositionAverager(4)

	// Add a sample
	sample := PositionSample{
		Latitude:   51.5074,
		Longitude:  -0.1278,
		Altitude:   45.0,
		FixQuality: 4,
		Timestamp:  time.Now().UTC(),
	}

	averager.AddSample(sample)

	// Reset the averager
	averager.Reset()

	// Now there should be no samples
	if averager.GetSampleCount() != 0 {
		t.Errorf("Expected 0 samples after reset, got %d", averager.GetSampleCount())
	}

	// The fix quality distribution should be reset
	if len(averager.fixQualityDist) != 0 {
		t.Errorf("Expected empty fix quality distribution after reset, got %d entries", len(averager.fixQualityDist))
	}
}

func TestGetFixQualityDistribution(t *testing.T) {
	averager := NewPositionAverager(4)

	// Add samples with different fix qualities
	samples := []PositionSample{
		{
			Latitude:   51.5074,
			Longitude:  -0.1278,
			Altitude:   45.0,
			FixQuality: 3,
			Timestamp:  time.Now().UTC(),
		},
		{
			Latitude:   51.5076,
			Longitude:  -0.1276,
			Altitude:   46.0,
			FixQuality: 4,
			Timestamp:  time.Now().UTC(),
		},
		{
			Latitude:   51.5078,
			Longitude:  -0.1274,
			Altitude:   47.0,
			FixQuality: 4,
			Timestamp:  time.Now().UTC(),
		},
		{
			Latitude:   51.5080,
			Longitude:  -0.1272,
			Altitude:   48.0,
			FixQuality: 5,
			Timestamp:  time.Now().UTC(),
		},
	}

	for _, sample := range samples {
		averager.AddSample(sample)
	}

	// Get the fix quality distribution
	dist := averager.GetFixQualityDistribution()

	// Check the distribution
	if dist[3] != 1 {
		t.Errorf("Expected 1 sample with fix quality 3, got %d", dist[3])
	}

	if dist[4] != 2 {
		t.Errorf("Expected 2 samples with fix quality 4, got %d", dist[4])
	}

	if dist[5] != 1 {
		t.Errorf("Expected 1 sample with fix quality 5, got %d", dist[5])
	}

	// Modify the returned distribution
	dist[3] = 100

	// The original distribution should not be affected
	if averager.fixQualityDist[3] != 1 {
		t.Errorf("Expected original distribution to be unchanged, got %d", averager.fixQualityDist[3])
	}
}
