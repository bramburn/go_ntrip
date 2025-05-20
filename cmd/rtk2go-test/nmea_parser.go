package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// NMEAParserImpl implements the NMEAParser interface
type NMEAParserImpl struct{}

// NewNMEAParser creates a new NMEA parser
func NewNMEAParser() *NMEAParserImpl {
	return &NMEAParserImpl{}
}

// Parse parses an NMEA sentence
func (p *NMEAParserImpl) Parse(sentence string) (NMEASentence, error) {
	result := NMEASentence{
		Raw:   sentence,
		Valid: false,
	}

	// Check for minimum length
	if len(sentence) < 6 {
		return result, fmt.Errorf("sentence too short")
	}

	// Check for valid start character
	if sentence[0] != '$' {
		return result, fmt.Errorf("invalid start character")
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
		return result, fmt.Errorf("not enough fields")
	}

	// Extract sentence type
	result.Type = strings.TrimPrefix(fields[0], "$")
	result.Fields = fields[1:]
	result.Valid = true

	return result, nil
}

// ParseGGA parses a GGA sentence
func (p *NMEAParserImpl) ParseGGA(sentence string) (GGAData, error) {
	var data GGAData

	// Parse the sentence first
	parsed, err := p.Parse(sentence)
	if err != nil {
		return data, err
	}

	if !parsed.Valid {
		return data, fmt.Errorf("invalid NMEA sentence")
	}

	// Check if it's a GGA sentence
	if !strings.HasSuffix(parsed.Type, "GGA") {
		return data, fmt.Errorf("not a GGA sentence")
	}

	// Check if we have enough fields
	if len(parsed.Fields) < 14 {
		return data, fmt.Errorf("not enough fields in GGA sentence")
	}

	// Parse time
	data.Time = parsed.Fields[0]

	// Parse latitude
	if parsed.Fields[1] != "" {
		lat, err := strconv.ParseFloat(parsed.Fields[1], 64)
		if err == nil {
			// Convert NMEA format (DDMM.MMMM) to decimal degrees
			latDeg := math.Floor(lat / 100.0)
			latMin := lat - latDeg*100.0
			data.Latitude = latDeg + latMin/60.0

			// Apply direction
			if parsed.Fields[2] == "S" {
				data.Latitude = -data.Latitude
			}
		}
	}
	data.LatDir = parsed.Fields[2]

	// Parse longitude
	if parsed.Fields[3] != "" {
		lon, err := strconv.ParseFloat(parsed.Fields[3], 64)
		if err == nil {
			// Convert NMEA format (DDDMM.MMMM) to decimal degrees
			lonDeg := math.Floor(lon / 100.0)
			lonMin := lon - lonDeg*100.0
			data.Longitude = lonDeg + lonMin/60.0

			// Apply direction
			if parsed.Fields[4] == "W" {
				data.Longitude = -data.Longitude
			}
		}
	}
	data.LonDir = parsed.Fields[4]

	// Parse fix quality
	if parsed.Fields[5] != "" {
		quality, err := strconv.Atoi(parsed.Fields[5])
		if err == nil {
			data.Quality = quality
		}
	}

	// Parse number of satellites
	if parsed.Fields[6] != "" {
		sats, err := strconv.Atoi(parsed.Fields[6])
		if err == nil {
			data.NumSats = sats
		}
	}

	// Parse HDOP
	if parsed.Fields[7] != "" {
		hdop, err := strconv.ParseFloat(parsed.Fields[7], 64)
		if err == nil {
			data.HDOP = hdop
		}
	}

	// Parse altitude
	if parsed.Fields[8] != "" {
		alt, err := strconv.ParseFloat(parsed.Fields[8], 64)
		if err == nil {
			data.Altitude = alt
		}
	}
	data.AltUnit = parsed.Fields[9]

	// Parse geoid separation
	if parsed.Fields[10] != "" {
		geoid, err := strconv.ParseFloat(parsed.Fields[10], 64)
		if err == nil {
			data.GeoidSep = geoid
		}
	}
	data.GeoidUnit = parsed.Fields[11]

	// Parse age of differential
	if parsed.Fields[12] != "" {
		age, err := strconv.ParseFloat(parsed.Fields[12], 64)
		if err == nil {
			data.DGPSAge = age
		}
	}

	// Parse DGPS station ID
	data.DGPSStaID = parsed.Fields[13]

	return data, nil
}

// GetFixQualityName returns a human-readable name for the fix quality
func GetFixQualityName(quality int) string {
	switch quality {
	case 0:
		return rtkStatusNone
	case 1:
		return rtkStatusSingle
	case 2:
		return rtkStatusDGPS
	case 4:
		return rtkStatusFix
	case 5:
		return rtkStatusFloat
	default:
		return rtkStatusNone
	}
}

// RTKSolution represents an RTK solution
type RTKSolution struct {
	Status    string    // Current RTK status (NONE, SINGLE, FLOAT, FIX)
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	HDOP      float64   // Horizontal dilution of precision
	Age       float64   // Age of differential corrections in seconds
	Time      time.Time // Time of the last update
}
