package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bramburn/go_ntrip/internal/ntrip"
	"github.com/bramburn/go_ntrip/internal/position"
	"github.com/bramburn/go_ntrip/internal/rtk"
)

func main() {
	// Parse command line flags
	address := flag.String("address", "", "NTRIP server address (e.g., 192.168.0.64)")
	port := flag.String("port", "2101", "NTRIP server port")
	username := flag.String("user", "", "Username for NTRIP server")
	password := flag.String("pass", "", "Password for NTRIP server")
	mountpoint := flag.String("mount", "", "Mountpoint name")
	outputFile := flag.String("output", "", "Output file path (default: ./rtk_position.json)")
	minFixQuality := flag.Int("min-fix", 4, "Minimum fix quality (4=RTK Fixed, 5=Float RTK)")
	sampleCount := flag.Int("samples", 60, "Number of samples to collect")
	timeout := flag.Duration("timeout", 10*time.Minute, "Timeout for connection")
	flag.Parse()

	// Check required parameters
	if *address == "" {
		fmt.Println("Error: NTRIP server address is required")
		flag.Usage()
		os.Exit(1)
	}

	if *mountpoint == "" {
		fmt.Println("Error: Mountpoint is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output file if not specified
	if *outputFile == "" {
		execPath, err := os.Executable()
		if err != nil {
			execPath = "."
		}
		*outputFile = filepath.Join(filepath.Dir(execPath), "rtk_position.json")
	}

	// Validate fix quality
	if *minFixQuality < 0 || *minFixQuality > 8 {
		fmt.Println("Error: Invalid fix quality. Must be between 0 and 8.")
		flag.Usage()
		os.Exit(1)
	}

	// Validate sample count
	if *sampleCount <= 0 {
		fmt.Println("Error: Invalid sample count. Must be greater than 0.")
		flag.Usage()
		os.Exit(1)
	}

	// Construct URL
	url := fmt.Sprintf("http://%s:%s", *address, *port)

	// Create NTRIP client
	client := ntrip.NewClient(url, *username, *password, *mountpoint)

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived shutdown signal")
		cancel()
	}()

	// Connect to NTRIP server
	fmt.Printf("Connecting to NTRIP server at %s...\n", url)
	stream, err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error connecting to NTRIP server: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	fmt.Println("Connected to NTRIP server.")
	fmt.Printf("Processing RTCM data (minimum fix quality: %s)...\n",
		position.GetFixQualityDescription(*minFixQuality))
	fmt.Printf("Will collect %d samples. Press Ctrl+C to stop early.\n", *sampleCount)

	// Create RTK processor
	processor := rtk.NewProcessor()

	// Create position averager
	averager := position.NewPositionAverager(*minFixQuality)

	// Start processing
	processor.StartProcessing()

	// Create a channel to signal completion
	doneChan := make(chan struct{})

	// Start goroutine to collect solutions
	go func() {
		solutionChan := processor.GetSolutionChannel()
		currentCount := 0

		for {
			select {
			case <-ctx.Done():
				doneChan <- struct{}{}
				return
			case solution := <-solutionChan:
				// Convert solution to position
				pos := solution.ToPosition()

				// Create sample
				sample := position.PositionSample{
					Latitude:   pos.Latitude,
					Longitude:  pos.Longitude,
					Altitude:   pos.Altitude,
					FixQuality: pos.FixQuality,
					Timestamp:  pos.Timestamp,
				}

				// Add sample to averager
				if averager.AddSample(sample) {
					// Increment count if sample was accepted
					currentCount++
					fmt.Printf("Sample %d/%d collected (Fix: %s)\r",
						currentCount, *sampleCount, position.GetFixQualityDescription(pos.FixQuality))

					// Check if we've collected enough samples
					if currentCount >= *sampleCount {
						fmt.Println("\nCollected requested number of samples.")
						doneChan <- struct{}{}
						return
					}
				} else {
					// Display current fix quality if sample was rejected
					fmt.Printf("Current fix quality: %s (not used)\r",
						position.GetFixQualityDescription(pos.FixQuality))
				}
			}
		}
	}()

	// Read RTCM data and feed it to the processor
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stream.Read(buffer)
				if err != nil {
					fmt.Printf("\nError reading from NTRIP stream: %v\n", err)
					doneChan <- struct{}{}
					return
				}
				if n > 0 {
					// Process RTCM data
					processor.ProcessRTCM(buffer[:n])
				}
			}
		}
	}()

	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		fmt.Println("\nTimeout or cancellation")
	case <-doneChan:
		fmt.Println("\nProcessing complete")
	}

	// Stop processing
	processor.StopProcessing()

	// Process results
	processResults(averager, *outputFile)
}

// processResults processes and displays the averaged position results
func processResults(averager *position.PositionAverager, outputFile string) {
	// Check if we have any samples
	if averager.GetSampleCount() > 0 {
		// Get averaged position
		pos, stats, err := averager.GetAveragedPosition()
		if err != nil {
			fmt.Printf("\nError getting averaged position: %v\n", err)
			return
		}

		// Display position information
		fmt.Println("\nAveraged position:")
		fmt.Printf("  Latitude: %.8f (±%.8f)\n", pos.Latitude, stats.LatitudeStdDev)
		fmt.Printf("  Longitude: %.8f (±%.8f)\n", pos.Longitude, stats.LongitudeStdDev)
		fmt.Printf("  Altitude: %.2f meters (±%.2f)\n", pos.Altitude, stats.AltitudeStdDev)
		fmt.Printf("  Sample Count: %d\n", stats.SampleCount)
		fmt.Printf("  Duration: %.1f seconds\n", stats.Duration)
		fmt.Printf("  Timestamp: %s\n", pos.Timestamp.Format(time.RFC3339))

		// Display fix quality distribution
		fmt.Println("  Fix Quality Distribution:")
		for quality, count := range stats.FixQualityDistribution {
			fmt.Printf("    %s: %d samples\n", position.GetFixQualityDescription(quality), count)
		}

		// Save position to file
		err = position.SavePositionWithStats(pos, stats, outputFile)
		if err != nil {
			fmt.Printf("Error saving position to file: %v\n", err)
		} else {
			fmt.Printf("Position saved to %s\n", outputFile)
		}
	} else {
		fmt.Println("\nNo position samples collected.")
	}
}
