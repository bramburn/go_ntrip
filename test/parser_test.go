package test

import (
	"testing"

	"github.com/bramburn/go_ntrip/internal/parser"
)

func TestNMEAParser(t *testing.T) {
	p := parser.NewNMEAParser()
	
	// Test valid NMEA sentence
	sentence := "$GNGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	parsed := p.Parse(sentence)
	
	if !parsed.Valid {
		t.Error("Expected valid NMEA sentence, got invalid")
	}
	
	if parsed.Type != "GNGGA" {
		t.Errorf("Expected type GNGGA, got %s", parsed.Type)
	}
	
	if len(parsed.Fields) != 14 {
		t.Errorf("Expected 14 fields, got %d", len(parsed.Fields))
	}
	
	if parsed.Fields[0] != "123519" {
		t.Errorf("Expected time 123519, got %s", parsed.Fields[0])
	}
	
	if parsed.Checksum != "47" {
		t.Errorf("Expected checksum 47, got %s", parsed.Checksum)
	}
	
	// Test invalid NMEA sentence
	invalidSentence := "INVALID"
	invalidParsed := p.Parse(invalidSentence)
	
	if invalidParsed.Valid {
		t.Error("Expected invalid NMEA sentence, got valid")
	}
}

func TestNMEAFormatting(t *testing.T) {
	p := parser.NewNMEAParser()
	
	// Test time formatting
	time := "123519.00"
	formattedTime := p.FormatTime(time)
	expected := "12:35:19.00"
	
	if formattedTime != expected {
		t.Errorf("Expected formatted time %s, got %s", expected, formattedTime)
	}
	
	// Test date formatting
	date := "230521"
	formattedDate := p.FormatDate(date)
	expectedDate := "23/05/2021"
	
	if formattedDate != expectedDate {
		t.Errorf("Expected formatted date %s, got %s", expectedDate, formattedDate)
	}
	
	// Test fix quality
	quality := "1"
	fixQuality := p.GetFixQuality(quality)
	expectedQuality := "GPS Fix (1)"
	
	if fixQuality != expectedQuality {
		t.Errorf("Expected fix quality %s, got %s", expectedQuality, fixQuality)
	}
}

func TestRTCMParser(t *testing.T) {
	p := parser.NewRTCMParser()
	
	// Create a simple RTCM message for testing
	// RTCM message with preamble 0xD3, length 10, type 1074
	data := []byte{0xD3, 0x00, 0x0A, 0x42, 0xA0, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B}
	
	messages := p.Process(data)
	
	if len(messages) != 1 {
		t.Errorf("Expected 1 RTCM message, got %d", len(messages))
		return
	}
	
	msg := messages[0]
	
	if !msg.Valid {
		t.Error("Expected valid RTCM message, got invalid")
	}
	
	if msg.MessageType != 1074 {
		t.Errorf("Expected message type 1074, got %d", msg.MessageType)
	}
	
	if msg.Length != 10 {
		t.Errorf("Expected length 10, got %d", msg.Length)
	}
	
	// Test message description
	description := p.GetMessageDescription(1074)
	expectedDesc := "GPS MSM4"
	
	if description != expectedDesc {
		t.Errorf("Expected description %s, got %s", expectedDesc, description)
	}
}

func TestUBXParser(t *testing.T) {
	p := parser.NewUBXParser()
	
	// Create a simple UBX message for testing
	// UBX message with header 0xB5 0x62, class 0x01, ID 0x07, length 4
	data := []byte{0xB5, 0x62, 0x01, 0x07, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04, 0xAA, 0xBB}
	
	messages := p.Process(data)
	
	if len(messages) != 1 {
		t.Errorf("Expected 1 UBX message, got %d", len(messages))
		return
	}
	
	msg := messages[0]
	
	if !msg.Valid {
		t.Error("Expected valid UBX message, got invalid")
	}
	
	if msg.Class != 0x01 {
		t.Errorf("Expected class 0x01, got 0x%02X", msg.Class)
	}
	
	if msg.ID != 0x07 {
		t.Errorf("Expected ID 0x07, got 0x%02X", msg.ID)
	}
	
	if msg.Length != 4 {
		t.Errorf("Expected length 4, got %d", msg.Length)
	}
	
	// Test class description
	classDesc := p.GetClassDescription(0x01)
	expectedClassDesc := "NAV (Navigation Results)"
	
	if classDesc != expectedClassDesc {
		t.Errorf("Expected class description %s, got %s", expectedClassDesc, classDesc)
	}
	
	// Test message description
	msgDesc := p.GetMessageDescription(0x01, 0x07)
	expectedMsgDesc := "NAV-PVT (Navigation Position Velocity Time Solution)"
	
	if msgDesc != expectedMsgDesc {
		t.Errorf("Expected message description %s, got %s", expectedMsgDesc, msgDesc)
	}
}
