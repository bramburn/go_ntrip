package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bramburn/go_ntrip/pkg/ntrip"
)

func main() {
	// Parse command line flags
	ntripServer := flag.String("server", "192.168.0.64", "NTRIP server address")
	ntripPort := flag.String("port", "2101", "NTRIP server port")
	ntripUser := flag.String("user", "reach", "NTRIP username")
	ntripPassword := flag.String("password", "emlidreach", "NTRIP password")
	ntripMountpoint := flag.String("mountpoint", "REACH", "NTRIP mountpoint")
	gnssPort := flag.String("gnss", "COM3:115200:8:N:1", "GNSS receiver port")
	duration := flag.Int("duration", 60, "Duration to run in seconds")
	flag.Parse()

	// Create a new NTRIP client
	client, err := ntrip.NewClient(*ntripServer, *ntripPort, *ntripUser, *ntripPassword, *ntripMountpoint)
	if err != nil {
		log.Fatalf("Failed to create NTRIP client: %v", err)
	}

	// Connect to the NTRIP server
	err = client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to NTRIP server: %v", err)
	}
	defer client.Disconnect()

	// Connect to the GNSS receiver
	receiver, err := ntrip.NewGNSSReceiver(*gnssPort)
	if err != nil {
		log.Fatalf("Failed to connect to GNSS receiver: %v", err)
	}
	defer receiver.Close()

	// Start the RTK processing
	processor, err := ntrip.NewRTKProcessor(receiver, client)
	if err != nil {
		log.Fatalf("Failed to create RTK processor: %v", err)
	}

	// Start processing
	err = processor.Start()
	if err != nil {
		log.Fatalf("Failed to start RTK processing: %v", err)
	}

	// Run for the specified duration
	fmt.Printf("RTK processing started. Running for %d seconds...\n", *duration)
	time.Sleep(time.Duration(*duration) * time.Second)

	// Stop processing
	processor.Stop()
	fmt.Println("RTK processing stopped.")

	// Print results
	stats := processor.GetStats()
	fmt.Printf("RTK processing statistics:\n")
	fmt.Printf("  Rover observations: %d\n", stats.RoverObs)
	fmt.Printf("  Base observations: %d\n", stats.BaseObs)
	fmt.Printf("  Solutions: %d\n", stats.Solutions)
	fmt.Printf("  Fix ratio: %.2f%%\n", stats.FixRatio*100)

	// Exit with success
	os.Exit(0)
}
