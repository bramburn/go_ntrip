package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Save original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save original flag.CommandLine
	oldCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldCommandLine }()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set up test arguments
	os.Args = []string{
		"cmd",
		"-address", "example.com",
		"-port", "2101",
		"-user", "testuser",
		"-pass", "testpass",
		"-mount", "TESTMOUNT",
		"-output", "test.json",
		"-min-fix", "5",
		"-samples", "30",
	}

	// Parse flags
	address := flag.String("address", "", "NTRIP server address")
	port := flag.String("port", "2101", "NTRIP server port")
	username := flag.String("user", "", "Username for NTRIP server")
	password := flag.String("pass", "", "Password for NTRIP server")
	mountpoint := flag.String("mount", "", "Mountpoint name")
	outputFile := flag.String("output", "", "Output file path")
	minFixQuality := flag.Int("min-fix", 4, "Minimum fix quality")
	sampleCount := flag.Int("samples", 60, "Number of samples to collect")
	flag.Parse()

	// Check parsed values
	if *address != "example.com" {
		t.Errorf("Expected address 'example.com', got '%s'", *address)
	}

	if *port != "2101" {
		t.Errorf("Expected port '2101', got '%s'", *port)
	}

	if *username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", *username)
	}

	if *password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", *password)
	}

	if *mountpoint != "TESTMOUNT" {
		t.Errorf("Expected mountpoint 'TESTMOUNT', got '%s'", *mountpoint)
	}

	if *outputFile != "test.json" {
		t.Errorf("Expected output file 'test.json', got '%s'", *outputFile)
	}

	if *minFixQuality != 5 {
		t.Errorf("Expected min fix quality 5, got %d", *minFixQuality)
	}

	if *sampleCount != 30 {
		t.Errorf("Expected sample count 30, got %d", *sampleCount)
	}
}
