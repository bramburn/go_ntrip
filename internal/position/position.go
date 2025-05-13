package position

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bramburn/go_ntrip/internal/parser"
)

// Position represents a GNSS position
type Position struct {
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Altitude    float64        `json:"altitude"`
	FixQuality  int            `json:"fix_quality"`
	Satellites  int            `json:"satellites"`
	HDOP        float64        `json:"hdop"`
	Timestamp   time.Time      `json:"timestamp"`
	Description string         `json:"description"`
	Stats       *PositionStats `json:"stats,omitempty"`
}

// ExtractFromGGA extracts position information from a GGA NMEA sentence
func ExtractFromGGA(sentence parser.NMEASentence) (*Position, error) {
	if !strings.HasSuffix(sentence.Type, "GGA") || len(sentence.Fields) < 14 {
		return nil, fmt.Errorf("not a valid GGA sentence")
	}

	// Parse time
	timeStr := sentence.Fields[0]
	var timestamp time.Time
	if len(timeStr) >= 6 {
		hour, _ := strconv.Atoi(timeStr[0:2])
		minute, _ := strconv.Atoi(timeStr[2:4])
		second, _ := strconv.ParseFloat(timeStr[4:], 64)
		timestamp = time.Now().UTC()
		timestamp = time.Date(
			timestamp.Year(), timestamp.Month(), timestamp.Day(),
			hour, minute, int(second), int((second-float64(int(second)))*1e9),
			time.UTC,
		)
	}

	// Parse latitude
	lat, _ := strconv.ParseFloat(sentence.Fields[1], 64)
	latDir := sentence.Fields[2]
	latitude := convertNMEACoordinate(lat, latDir == "S")

	// Parse longitude
	lon, _ := strconv.ParseFloat(sentence.Fields[3], 64)
	lonDir := sentence.Fields[4]
	longitude := convertNMEACoordinate(lon, lonDir == "W")

	// Parse fix quality
	fixQuality, _ := strconv.Atoi(sentence.Fields[5])

	// Parse satellites
	satellites, _ := strconv.Atoi(sentence.Fields[6])

	// Parse HDOP
	hdop, _ := strconv.ParseFloat(sentence.Fields[7], 64)

	// Parse altitude
	altitude, _ := strconv.ParseFloat(sentence.Fields[8], 64)

	return &Position{
		Latitude:    latitude,
		Longitude:   longitude,
		Altitude:    altitude,
		FixQuality:  fixQuality,
		Satellites:  satellites,
		HDOP:        hdop,
		Timestamp:   timestamp,
		Description: getFixQualityDescription(fixQuality),
	}, nil
}

// SaveToFile saves the position to a JSON file
func (p *Position) SaveToFile(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

// SavePositionWithStats saves a position with stats to a JSON file
func SavePositionWithStats(pos *Position, stats *PositionStats, filePath string) error {
	// Attach stats to position
	pos.Stats = stats

	// Save to file
	return pos.SaveToFile(filePath)
}

// LoadFromFile loads a position from a JSON file
func LoadFromFile(filePath string) (*Position, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal from JSON
	var position Position
	if err := json.Unmarshal(data, &position); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &position, nil
}

// convertNMEACoordinate converts NMEA coordinate format (DDMM.MMMM) to decimal degrees
func convertNMEACoordinate(coord float64, isNegative bool) float64 {
	// Extract degrees and minutes
	degrees := float64(int(coord / 100))
	minutes := coord - degrees*100

	// Convert to decimal degrees
	decimal := degrees + minutes/60.0

	// Apply sign
	if isNegative {
		decimal = -decimal
	}

	return decimal
}

// getFixQualityDescription returns a description of the fix quality
func getFixQualityDescription(quality int) string {
	return GetFixQualityDescription(quality)
}

// GetFixQualityDescription returns a description of the fix quality (exported version)
func GetFixQualityDescription(quality int) string {
	switch quality {
	case 0:
		return "Invalid"
	case 1:
		return "GPS Fix"
	case 2:
		return "DGPS Fix"
	case 3:
		return "PPS Fix"
	case 4:
		return "RTK Fix"
	case 5:
		return "Float RTK"
	case 6:
		return "Estimated"
	case 7:
		return "Manual Input"
	case 8:
		return "Simulation"
	default:
		return fmt.Sprintf("Unknown (%d)", quality)
	}
}
