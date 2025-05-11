package parser

import (
	"strings"
)

// NMEASentence represents a parsed NMEA sentence
type NMEASentence struct {
	Type     string   // Sentence type (e.g., "GNGGA", "GNRMC")
	Fields   []string // Data fields
	Valid    bool     // Whether the sentence is valid
	Checksum string   // Checksum value
}

// NMEAParser provides functionality to parse NMEA sentences
type NMEAParser struct{}

// NewNMEAParser creates a new NMEA parser
func NewNMEAParser() *NMEAParser {
	return &NMEAParser{}
}

// Parse parses an NMEA sentence
func (p *NMEAParser) Parse(sentence string) NMEASentence {
	result := NMEASentence{
		Valid: false,
	}

	// Check for minimum length
	if len(sentence) < 6 {
		return result
	}

	// Check for valid start character
	if sentence[0] != '$' {
		return result
	}

	// Extract checksum if present
	checksumPos := strings.LastIndex(sentence, "*")
	var data string
	if checksumPos != -1 && checksumPos < len(sentence)-2 {
		data = sentence[:checksumPos]
		result.Checksum = sentence[checksumPos+1:]
	} else {
		data = sentence
	}

	// Split into fields
	fields := strings.Split(data, ",")
	if len(fields) < 2 {
		return result
	}

	// Extract sentence type
	result.Type = strings.TrimPrefix(fields[0], "$")
	result.Fields = fields[1:]
	result.Valid = true

	return result
}

// FormatTime formats NMEA time string (HHMMSS.sss)
func (p *NMEAParser) FormatTime(timeStr string) string {
	if len(timeStr) < 6 {
		return timeStr
	}

	hours := timeStr[0:2]
	minutes := timeStr[2:4]
	seconds := timeStr[4:]

	return hours + ":" + minutes + ":" + seconds
}

// FormatDate formats NMEA date string (DDMMYY)
func (p *NMEAParser) FormatDate(dateStr string) string {
	if len(dateStr) != 6 {
		return dateStr
	}

	day := dateStr[0:2]
	month := dateStr[2:4]
	year := dateStr[4:6]

	return day + "/" + month + "/20" + year
}

// FormatLatLon formats latitude/longitude from NMEA format
func (p *NMEAParser) FormatLatLon(coord string) string {
	if coord == "" {
		return "N/A"
	}
	return coord
}

// GetFixQuality returns a description of the fix quality
func (p *NMEAParser) GetFixQuality(quality string) string {
	switch quality {
	case "0":
		return "Invalid (0)"
	case "1":
		return "GPS Fix (1)"
	case "2":
		return "DGPS Fix (2)"
	case "3":
		return "PPS Fix (3)"
	case "4":
		return "RTK Fix (4)"
	case "5":
		return "Float RTK (5)"
	case "6":
		return "Estimated (6)"
	case "7":
		return "Manual Input (7)"
	case "8":
		return "Simulation (8)"
	default:
		return quality
	}
}

// GetFixType returns a description of the fix type
func (p *NMEAParser) GetFixType(fixType string) string {
	switch fixType {
	case "1":
		return "No Fix (1)"
	case "2":
		return "2D Fix (2)"
	case "3":
		return "3D Fix (3)"
	default:
		return fixType
	}
}
