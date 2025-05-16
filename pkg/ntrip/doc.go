/*
Package ntrip provides functionality for working with NTRIP (Networked Transport of RTCM via Internet Protocol)
and GNSS (Global Navigation Satellite System) data processing.

This package serves as a high-level wrapper around the core functionality provided by the gnssgo package,
making it easier to work with NTRIP clients, GNSS receivers, and RTK (Real-Time Kinematic) processing.

# Main Components

## NTRIP Client

The Client type provides a simple interface for connecting to NTRIP servers and receiving RTCM correction data.
It handles the details of establishing and maintaining the connection, authentication, and data streaming.

Example usage:
    
    // Create a new NTRIP client
    client, err := ntrip.NewClient("example.com", "2101", "username", "password", "MOUNTPOINT")
    if err != nil {
        log.Fatalf("Failed to create NTRIP client: %v", err)
    }
    
    // Connect to the NTRIP server
    err = client.Connect()
    if err != nil {
        log.Fatalf("Failed to connect to NTRIP server: %v", err)
    }
    defer client.Disconnect()
    
    // Read RTCM data
    buffer := make([]byte, 1024)
    n, err := client.Read(buffer)
    if err != nil {
        log.Fatalf("Failed to read RTCM data: %v", err)
    }
    
    // Process the RTCM data
    processRTCM(buffer[:n])

## GNSS Receiver

The GNSSReceiver type provides an interface for connecting to physical GNSS receivers via serial ports
and reading raw GNSS data. It handles the details of establishing and maintaining the connection and data streaming.

Example usage:
    
    // Create a new GNSS receiver
    receiver, err := ntrip.NewGNSSReceiver("COM1:9600:8:N:1")
    if err != nil {
        log.Fatalf("Failed to connect to GNSS receiver: %v", err)
    }
    defer receiver.Close()
    
    // Read GNSS data
    buffer := make([]byte, 1024)
    n, err := receiver.Read(buffer)
    if err != nil {
        log.Fatalf("Failed to read GNSS data: %v", err)
    }
    
    // Process the GNSS data
    processGNSS(buffer[:n])

## RTK Processor

The RTKProcessor type provides functionality for processing GNSS data using RTK techniques.
It combines data from a GNSS receiver (rover) and NTRIP client (base station) to calculate precise positions.

Example usage:
    
    // Create a new RTK processor
    processor, err := ntrip.NewRTKProcessor(receiver, client)
    if err != nil {
        log.Fatalf("Failed to create RTK processor: %v", err)
    }
    
    // Start RTK processing
    err = processor.Start()
    if err != nil {
        log.Fatalf("Failed to start RTK processing: %v", err)
    }
    defer processor.Stop()
    
    // Get RTK statistics
    stats := processor.GetStats()
    fmt.Printf("Rover observations: %d\n", stats.RoverObs)
    fmt.Printf("Base observations: %d\n", stats.BaseObs)
    fmt.Printf("Solutions: %d\n", stats.Solutions)
    fmt.Printf("Fix ratio: %.2f%%\n", stats.FixRatio*100)

# Relationship with gnssgo

This package is built on top of the gnssgo package, which provides the core GNSS processing functionality.
The ntrip package provides a higher-level, more user-friendly interface to the gnssgo functionality,
making it easier to work with NTRIP and RTK in Go applications.
*/
package ntrip
