package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RTK status constants
const (
	rtkStatusNone   = "NONE"   // No position
	rtkStatusSingle = "SINGLE" // Single solution
	rtkStatusDGPS   = "DGPS"   // DGPS solution
	rtkStatusFloat  = "FLOAT"  // Float solution
	rtkStatusFix    = "FIX"    // Fixed solution
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

func main() {
	// Parse command line flags
	ntripServer := flag.String("server", "rtk2go.com", "NTRIP server address")
	ntripPort := flag.String("port", "2101", "NTRIP server port")
	ntripUser := flag.String("user", "nitrogen@gmail.com", "NTRIP username (email address)")
	ntripPassword := flag.String("password", "password", "NTRIP password (any value for RTK2go)")
	ntripMountpoint := flag.String("mountpoint", "OCF-RH55LS-Capel", "NTRIP mountpoint (OCF-RH55LS-Capel, MEDW, ozzy1)")
	gnssPort := flag.String("gnss", "COM3", "GNSS receiver port")
	baudRate := flag.Int("baud", 38400, "GNSS receiver baud rate")
	duration := flag.Int("duration", 0, "Duration to run in seconds (0 for indefinite)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	colorOutput := flag.Bool("color", true, "Enable colored output for RTK status")
	reconnect := flag.Bool("reconnect", true, "Automatically reconnect on connection loss")
	reconnectInterval := flag.Int("reconnect-interval", 5, "Reconnection interval in seconds")
	flag.Parse()

	// List available mountpoints
	availableMountpoints := []string{"OCF-RH55LS-Capel", "MEDW", "ozzy1"}

	// Validate mountpoint
	validMountpoint := false
	for _, m := range availableMountpoints {
		if *ntripMountpoint == m {
			validMountpoint = true
			break
		}
	}

	if !validMountpoint {
		fmt.Printf("Warning: Mountpoint '%s' is not in the list of known mountpoints: %v\n",
			*ntripMountpoint, availableMountpoints)
		fmt.Println("Continuing anyway, but connection might fail...")
	}

	// Set up logging
	log.SetFlags(log.Ltime | log.Ldate | log.Lshortfile)

	// Create a console logger
	consoleLogger := log.New(os.Stdout, "", log.Ltime)

	// Print connection information
	consoleLogger.Printf("RTK2go Test Client")
	consoleLogger.Printf("NTRIP Server: %s:%s", *ntripServer, *ntripPort)
	consoleLogger.Printf("NTRIP Mountpoint: %s", *ntripMountpoint)
	consoleLogger.Printf("GNSS Receiver: %s:%d", *gnssPort, *baudRate)

	// Create the RTK application with options
	app := NewRTKApp(RTKAppOptions{
		ColorOutput:       *colorOutput,
		Reconnect:         *reconnect,
		ReconnectInterval: *reconnectInterval,
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		consoleLogger.Println("\nReceived shutdown signal")
		close(app.stopChan)
	}()

	// Connect to the GNSS receiver first
	consoleLogger.Printf("Connecting to GNSS receiver on port %s...\n", *gnssPort)

	// Create a new TOP708 receiver
	receiver, err := NewTOP708Receiver(*gnssPort, *baudRate)
	if err != nil {
		consoleLogger.Fatalf("Failed to create GNSS receiver: %v", err)
	}

	// Connect to the device
	err = receiver.Connect()
	if err != nil {
		consoleLogger.Fatalf("Failed to connect to device: %v", err)
	}
	defer receiver.Disconnect()

	// Verify the connection
	if !receiver.VerifyConnection(5 * time.Second) {
		consoleLogger.Fatalf("Failed to verify connection")
	}

	// Set the GNSS device in the app
	app.SetGNSSDevice(receiver)

	consoleLogger.Println("Connected to GNSS receiver successfully.")

	// Connect to the NTRIP server
	consoleLogger.Printf("Connecting to NTRIP server %s:%s...\n", *ntripServer, *ntripPort)
	
	// Create a new NTRIP client
	ntripClient, err := CreateNTRIPClient(*ntripServer, *ntripPort, *ntripUser, *ntripPassword, *ntripMountpoint)
	if err != nil {
		consoleLogger.Fatalf("Failed to create NTRIP client: %v", err)
	}

	// Set the NTRIP client in the app
	app.SetNTRIPClient(ntripClient)

	// Try to connect to NTRIP server
	err = ntripClient.Connect()
	if err != nil {
		if app.reconnect {
			consoleLogger.Printf("Failed to connect to NTRIP server: %v", err)
			consoleLogger.Printf("Will retry connection every %d seconds...", app.reconnectInterval)
		} else {
			consoleLogger.Fatalf("Failed to connect to NTRIP server: %v", err)
		}
	} else {
		consoleLogger.Println("Connected to NTRIP server successfully.")
	}
	// We'll handle disconnection manually to avoid panic
	// DO NOT call Disconnect() on the NTRIP client because it causes a panic
	// The OS will clean up the resources when the process exits

	// Start the RTK processor
	consoleLogger.Println("Starting RTK processor...")

	// Create a new RTK processor
	rtkProcessor, err := CreateRTKProcessor(receiver, ntripClient)
	if err != nil {
		consoleLogger.Fatalf("Failed to create RTK processor: %v", err)
	}

	// Set the RTK processor in the app
	app.SetRTKProcessor(rtkProcessor)

	// Start the RTK processor but don't panic if it fails
	err = rtkProcessor.Start()
	if err != nil {
		consoleLogger.Printf("Warning: Failed to start RTK processing: %v", err)
		consoleLogger.Println("Continuing without RTK processing...")
	} else {
		consoleLogger.Println("RTK processor started successfully.")
	}

	// Start the RTK application
	err = app.Start(consoleLogger)
	if err != nil {
		consoleLogger.Fatalf("Failed to start RTK application: %v", err)
	}

	// Run for the specified duration or until interrupted
	if *duration > 0 {
		select {
		case <-time.After(time.Duration(*duration) * time.Second):
			consoleLogger.Printf("Duration of %d seconds reached, shutting down...", *duration)
		case <-app.stopChan:
			// Shutdown signal received
		}
	} else {
		<-app.stopChan
	}

	// Stop the RTK application
	app.Stop(consoleLogger)
}
