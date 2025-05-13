package position

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// PositionSample represents a single position sample
type PositionSample struct {
	Latitude   float64
	Longitude  float64
	Altitude   float64
	FixQuality int
	Timestamp  time.Time
}

// PositionStats contains statistics about the averaged position
type PositionStats struct {
	SampleCount            int         `json:"sample_count"`
	Duration               float64     `json:"duration_seconds"`
	LatitudeStdDev         float64     `json:"latitude_std_dev"`
	LongitudeStdDev        float64     `json:"longitude_std_dev"`
	AltitudeStdDev         float64     `json:"altitude_std_dev"`
	StartTime              time.Time   `json:"start_time"`
	EndTime                time.Time   `json:"end_time"`
	FixQualityDistribution map[int]int `json:"fix_quality_distribution"`
}

// PositionAverager collects and averages position samples
type PositionAverager struct {
	samples        []PositionSample
	mutex          sync.Mutex
	minFixQuality  int
	startTime      time.Time
	fixQualityDist map[int]int
}

// NewPositionAverager creates a new position averager
func NewPositionAverager(minFixQuality int) *PositionAverager {
	return &PositionAverager{
		samples:        []PositionSample{},
		minFixQuality:  minFixQuality,
		startTime:      time.Now(),
		fixQualityDist: make(map[int]int),
	}
}

// AddSample adds a position sample to the averager
func (a *PositionAverager) AddSample(sample PositionSample) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Track fix quality distribution regardless of whether we use the sample
	a.fixQualityDist[sample.FixQuality]++

	// Only use samples with sufficient fix quality
	if sample.FixQuality < a.minFixQuality {
		return false
	}

	a.samples = append(a.samples, sample)
	return true
}

// GetSampleCount returns the number of samples collected
func (a *PositionAverager) GetSampleCount() int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return len(a.samples)
}

// GetAveragedPosition calculates the averaged position
func (a *PositionAverager) GetAveragedPosition() (*Position, *PositionStats, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.samples) == 0 {
		return nil, nil, fmt.Errorf("no samples collected")
	}

	// Calculate averages
	var sumLat, sumLon, sumAlt float64
	var minTime, maxTime time.Time

	// First pass: calculate sums
	for i, sample := range a.samples {
		sumLat += sample.Latitude
		sumLon += sample.Longitude
		sumAlt += sample.Altitude

		// Track min/max time
		if i == 0 || sample.Timestamp.Before(minTime) {
			minTime = sample.Timestamp
		}
		if i == 0 || sample.Timestamp.After(maxTime) {
			maxTime = sample.Timestamp
		}
	}

	avgLat := sumLat / float64(len(a.samples))
	avgLon := sumLon / float64(len(a.samples))
	avgAlt := sumAlt / float64(len(a.samples))

	// Second pass: calculate standard deviations
	var sumSqDiffLat, sumSqDiffLon, sumSqDiffAlt float64
	for _, sample := range a.samples {
		sumSqDiffLat += math.Pow(sample.Latitude-avgLat, 2)
		sumSqDiffLon += math.Pow(sample.Longitude-avgLon, 2)
		sumSqDiffAlt += math.Pow(sample.Altitude-avgAlt, 2)
	}

	// Calculate standard deviations
	stdDevLat := math.Sqrt(sumSqDiffLat / float64(len(a.samples)))
	stdDevLon := math.Sqrt(sumSqDiffLon / float64(len(a.samples)))
	stdDevAlt := math.Sqrt(sumSqDiffAlt / float64(len(a.samples)))

	// Create position object
	pos := &Position{
		Latitude:    avgLat,
		Longitude:   avgLon,
		Altitude:    avgAlt,
		FixQuality:  a.minFixQuality, // Use the minimum fix quality we accepted
		Satellites:  0,               // Not tracked in averager
		HDOP:        0,               // Not tracked in averager
		Timestamp:   time.Now().UTC(),
		Description: fmt.Sprintf("Averaged position from %d samples", len(a.samples)),
	}

	// Create stats object
	stats := &PositionStats{
		SampleCount:            len(a.samples),
		Duration:               maxTime.Sub(minTime).Seconds(),
		LatitudeStdDev:         stdDevLat,
		LongitudeStdDev:        stdDevLon,
		AltitudeStdDev:         stdDevAlt,
		StartTime:              minTime,
		EndTime:                maxTime,
		FixQualityDistribution: a.fixQualityDist,
	}

	return pos, stats, nil
}

// Reset clears all collected samples
func (a *PositionAverager) Reset() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.samples = []PositionSample{}
	a.startTime = time.Now()
	a.fixQualityDist = make(map[int]int)
}

// GetFixQualityDistribution returns the distribution of fix qualities
func (a *PositionAverager) GetFixQualityDistribution() map[int]int {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Create a copy to avoid race conditions
	dist := make(map[int]int)
	for k, v := range a.fixQualityDist {
		dist[k] = v
	}

	return dist
}
