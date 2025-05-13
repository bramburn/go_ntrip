package position

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bramburn/go_ntrip/internal/parser"
)

func TestExtractFromGGA(t *testing.T) {
	// Create a valid GGA sentence
	sentence := parser.NMEASentence{
		Type:     "GNGGA",
		Fields:   []string{"120000.00", "5130.44", "N", "00115.67", "E", "4", "10", "0.8", "100.0", "M", "0.0", "M", "", ""},
		Checksum: "7D",
		Valid:    true,
	}

	// Extract position
	pos, err := ExtractFromGGA(sentence)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check position
	if pos == nil {
		t.Fatal("Expected non-nil position")
	}

	// Check latitude (51°30.44' N = 51.5073333...)
	expectedLat := 51.50733333333333
	if pos.Latitude < expectedLat-0.0001 || pos.Latitude > expectedLat+0.0001 {
		t.Errorf("Expected latitude around %f, got %f", expectedLat, pos.Latitude)
	}

	// Check longitude (1°15.67' E = 1.2611666...)
	expectedLon := 1.2611666666666666
	if pos.Longitude < expectedLon-0.0001 || pos.Longitude > expectedLon+0.0001 {
		t.Errorf("Expected longitude around %f, got %f", expectedLon, pos.Longitude)
	}

	// Check altitude
	if pos.Altitude != 100.0 {
		t.Errorf("Expected altitude 100.0, got %f", pos.Altitude)
	}

	// Check fix quality
	if pos.FixQuality != 4 {
		t.Errorf("Expected fix quality 4, got %d", pos.FixQuality)
	}

	// Check satellites
	if pos.Satellites != 10 {
		t.Errorf("Expected 10 satellites, got %d", pos.Satellites)
	}

	// Check HDOP
	if pos.HDOP != 0.8 {
		t.Errorf("Expected HDOP 0.8, got %f", pos.HDOP)
	}

	// Check description
	if pos.Description != "RTK Fix" {
		t.Errorf("Expected description 'RTK Fix', got '%s'", pos.Description)
	}
}

func TestExtractFromGGAInvalid(t *testing.T) {
	// Test with non-GGA sentence
	sentence := parser.NMEASentence{
		Type:     "GNRMC",
		Fields:   []string{"120000.00", "A", "5130.44", "N", "00115.67", "E", "0.0", "0.0", "010122", "", "", "A"},
		Checksum: "7D",
		Valid:    true,
	}

	_, err := ExtractFromGGA(sentence)
	if err == nil {
		t.Error("Expected error with non-GGA sentence")
	}

	// Test with too few fields
	sentence = parser.NMEASentence{
		Type:     "GNGGA",
		Fields:   []string{"120000.00", "5130.44", "N"},
		Checksum: "7D",
		Valid:    true,
	}

	_, err = ExtractFromGGA(sentence)
	if err == nil {
		t.Error("Expected error with too few fields")
	}
}

func TestSaveToFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "position_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a position
	pos := &Position{
		Latitude:    51.5074,
		Longitude:   -0.1278,
		Altitude:    45.0,
		FixQuality:  4,
		Satellites:  10,
		HDOP:        0.8,
		Timestamp:   time.Now().UTC(),
		Description: "Test position",
	}

	// Save to file
	filePath := filepath.Join(tempDir, "position.json")
	err = pos.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save position: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to exist")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Parse JSON
	var loadedPos Position
	err = json.Unmarshal(data, &loadedPos)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check position
	if loadedPos.Latitude != pos.Latitude {
		t.Errorf("Expected latitude %f, got %f", pos.Latitude, loadedPos.Latitude)
	}

	if loadedPos.Longitude != pos.Longitude {
		t.Errorf("Expected longitude %f, got %f", pos.Longitude, loadedPos.Longitude)
	}

	if loadedPos.Altitude != pos.Altitude {
		t.Errorf("Expected altitude %f, got %f", pos.Altitude, loadedPos.Altitude)
	}

	if loadedPos.FixQuality != pos.FixQuality {
		t.Errorf("Expected fix quality %d, got %d", pos.FixQuality, loadedPos.FixQuality)
	}

	if loadedPos.Satellites != pos.Satellites {
		t.Errorf("Expected satellites %d, got %d", pos.Satellites, loadedPos.Satellites)
	}

	if loadedPos.HDOP != pos.HDOP {
		t.Errorf("Expected HDOP %f, got %f", pos.HDOP, loadedPos.HDOP)
	}

	if loadedPos.Description != pos.Description {
		t.Errorf("Expected description '%s', got '%s'", pos.Description, loadedPos.Description)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "position_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a position
	pos := &Position{
		Latitude:    51.5074,
		Longitude:   -0.1278,
		Altitude:    45.0,
		FixQuality:  4,
		Satellites:  10,
		HDOP:        0.8,
		Timestamp:   time.Now().UTC(),
		Description: "Test position",
	}

	// Save to file
	filePath := filepath.Join(tempDir, "position.json")
	err = pos.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save position: %v", err)
	}

	// Load from file
	loadedPos, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load position: %v", err)
	}

	// Check position
	if loadedPos.Latitude != pos.Latitude {
		t.Errorf("Expected latitude %f, got %f", pos.Latitude, loadedPos.Latitude)
	}

	if loadedPos.Longitude != pos.Longitude {
		t.Errorf("Expected longitude %f, got %f", pos.Longitude, loadedPos.Longitude)
	}

	if loadedPos.Altitude != pos.Altitude {
		t.Errorf("Expected altitude %f, got %f", pos.Altitude, loadedPos.Altitude)
	}

	if loadedPos.FixQuality != pos.FixQuality {
		t.Errorf("Expected fix quality %d, got %d", pos.FixQuality, loadedPos.FixQuality)
	}

	if loadedPos.Satellites != pos.Satellites {
		t.Errorf("Expected satellites %d, got %d", pos.Satellites, loadedPos.Satellites)
	}

	if loadedPos.HDOP != pos.HDOP {
		t.Errorf("Expected HDOP %f, got %f", pos.HDOP, loadedPos.HDOP)
	}

	if loadedPos.Description != pos.Description {
		t.Errorf("Expected description '%s', got '%s'", pos.Description, loadedPos.Description)
	}
}

func TestLoadFromFileError(t *testing.T) {
	// Test with non-existent file
	_, err := LoadFromFile("non_existent_file.json")
	if err == nil {
		t.Error("Expected error with non-existent file")
	}

	// Test with invalid JSON
	tempDir, err := os.MkdirTemp("", "position_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(filePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = LoadFromFile(filePath)
	if err == nil {
		t.Error("Expected error with invalid JSON")
	}
}

func TestSavePositionWithStats(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "position_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a position
	pos := &Position{
		Latitude:    51.5074,
		Longitude:   -0.1278,
		Altitude:    45.0,
		FixQuality:  4,
		Satellites:  10,
		HDOP:        0.8,
		Timestamp:   time.Now().UTC(),
		Description: "Test position",
	}

	// Create stats
	stats := &PositionStats{
		SampleCount:            3,
		Duration:               10.0,
		LatitudeStdDev:         0.0001,
		LongitudeStdDev:        0.0001,
		AltitudeStdDev:         0.1,
		StartTime:              time.Now().UTC().Add(-10 * time.Second),
		EndTime:                time.Now().UTC(),
		FixQualityDistribution: map[int]int{4: 2, 5: 1},
	}

	// Save position with stats
	filePath := filepath.Join(tempDir, "position_with_stats.json")
	err = SavePositionWithStats(pos, stats, filePath)
	if err != nil {
		t.Fatalf("Failed to save position with stats: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to exist")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Parse JSON
	var loadedPos Position
	err = json.Unmarshal(data, &loadedPos)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check position
	if loadedPos.Latitude != pos.Latitude {
		t.Errorf("Expected latitude %f, got %f", pos.Latitude, loadedPos.Latitude)
	}

	// Check stats
	if loadedPos.Stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if loadedPos.Stats.SampleCount != stats.SampleCount {
		t.Errorf("Expected sample count %d, got %d", stats.SampleCount, loadedPos.Stats.SampleCount)
	}

	if loadedPos.Stats.Duration != stats.Duration {
		t.Errorf("Expected duration %f, got %f", stats.Duration, loadedPos.Stats.Duration)
	}

	if loadedPos.Stats.LatitudeStdDev != stats.LatitudeStdDev {
		t.Errorf("Expected latitude std dev %f, got %f", stats.LatitudeStdDev, loadedPos.Stats.LatitudeStdDev)
	}

	if loadedPos.Stats.LongitudeStdDev != stats.LongitudeStdDev {
		t.Errorf("Expected longitude std dev %f, got %f", stats.LongitudeStdDev, loadedPos.Stats.LongitudeStdDev)
	}

	if loadedPos.Stats.AltitudeStdDev != stats.AltitudeStdDev {
		t.Errorf("Expected altitude std dev %f, got %f", stats.AltitudeStdDev, loadedPos.Stats.AltitudeStdDev)
	}

	// Check fix quality distribution
	if loadedPos.Stats.FixQualityDistribution[4] != 2 {
		t.Errorf("Expected 2 samples with fix quality 4, got %d", loadedPos.Stats.FixQualityDistribution[4])
	}

	if loadedPos.Stats.FixQualityDistribution[5] != 1 {
		t.Errorf("Expected 1 sample with fix quality 5, got %d", loadedPos.Stats.FixQualityDistribution[5])
	}
}

func TestConvertNMEACoordinate(t *testing.T) {
	// Test with positive latitude (DDMM.MMMM format)
	lat := 5130.44 // 51°30.44' N
	convertedLat := convertNMEACoordinate(lat, false)
	expectedLat := 51.50733333333333
	if convertedLat < expectedLat-0.0001 || convertedLat > expectedLat+0.0001 {
		t.Errorf("Expected latitude around %f, got %f", expectedLat, convertedLat)
	}

	// Test with negative latitude
	lat = 5130.44 // 51°30.44' S
	convertedLat = convertNMEACoordinate(lat, true)
	expectedLat = -51.50733333333333
	if convertedLat < expectedLat-0.0001 || convertedLat > expectedLat+0.0001 {
		t.Errorf("Expected latitude around %f, got %f", expectedLat, convertedLat)
	}

	// Test with positive longitude
	lon := 115.67 // 1°15.67' E
	convertedLon := convertNMEACoordinate(lon, false)
	expectedLon := 1.2611666666666667 // 1 + 15.67/60
	if convertedLon < expectedLon-0.0001 || convertedLon > expectedLon+0.0001 {
		t.Errorf("Expected longitude around %f, got %f", expectedLon, convertedLon)
	}

	// Test with negative longitude
	lon = 115.67 // 1°15.67' W
	convertedLon = convertNMEACoordinate(lon, true)
	expectedLon = -1.2611666666666667 // -(1 + 15.67/60)
	if convertedLon < expectedLon-0.0001 || convertedLon > expectedLon+0.0001 {
		t.Errorf("Expected longitude around %f, got %f", expectedLon, convertedLon)
	}
}

func TestGetFixQualityDescription(t *testing.T) {
	tests := []struct {
		quality  int
		expected string
	}{
		{0, "Invalid"},
		{1, "GPS Fix"},
		{2, "DGPS Fix"},
		{3, "PPS Fix"},
		{4, "RTK Fix"},
		{5, "Float RTK"},
		{6, "Estimated"},
		{7, "Manual Input"},
		{8, "Simulation"},
		{9, "Unknown (9)"},
	}

	for _, test := range tests {
		result := GetFixQualityDescription(test.quality)
		if result != test.expected {
			t.Errorf("Expected description '%s' for quality %d, got '%s'", test.expected, test.quality, result)
		}
	}
}
