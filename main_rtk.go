package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/bramburn/go_ntrip/pkg/ntrip"
	"github.com/go-gnss/rtcm"
	"go.bug.st/serial/enumerator"
)

// Configuration for the serial connection
const (
	defaultBaudRate = 38400 // Default baud rate for TOPGNSS TOP708
	defaultPort     = ""    // Leave empty to prompt user for port selection
	readTimeout     = 500 * time.Millisecond

	// Protocol identifiers
	protocolNMEA = "NMEA-0183"
	protocolRTCM = "RTCM3.3"
	protocolUBX  = "UBX"

	// Log file
	logFileName = "rtk.log"

	// RTK status constants
	rtkStatusNone   = 0
	rtkStatusSingle = 1
	rtkStatusFloat  = 2
	rtkStatusFix    = 4
)

// RTKStatus represents the current RTK status
type RTKStatus struct {
	Status    int       // RTK status (NONE, SINGLE, FLOAT, FIX)
	Time      time.Time // Time of the status
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	HDOP      float64   // Horizontal dilution of precision
	Age       float64   // Age of differential (seconds)
	Accuracy  float64   // Estimated horizontal accuracy (meters)
}

// RTCMFilter is a function that filters RTCM messages
type RTCMFilter func(msg rtcm.Message) bool

// RTKApp represents the RTK application
type RTKApp struct {
	ntripClient  *ntrip.Client
	gnssReceiver *ntrip.GNSSReceiver
	rtkProcessor *ntrip.RTKProcessor
	logger       *log.Logger
	status       RTKStatus
	statusMutex  sync.Mutex
	rtcmFilters  []RTCMFilter
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func main() {
	// Parse command line flags
	ntripServer := flag.String("server", "192.168.0.64", "NTRIP server address")
	ntripPort := flag.String("port", "2101", "NTRIP server port")
	ntripUser := flag.String("user", "reach", "NTRIP username")
	ntripPassword := flag.String("password", "emlidreach", "NTRIP password")
	ntripMountpoint := flag.String("mountpoint", "REACH", "NTRIP mountpoint")
	gnssPort := flag.String("gnss", "", "GNSS receiver port (leave empty to select)")
	logFile := flag.String("log", logFileName, "Log file name")
	flag.Parse()

	// Setup logging
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)

	// Create a multi-writer to log to both file and console
	multiWriter := io.MultiWriter(os.Stdout, f)
	consoleLogger := log.New(multiWriter, "", log.LstdFlags)

	consoleLogger.Println("GNSS RTK Position Calculator")
	consoleLogger.Println("---------------------------")

	// Select GNSS port if not specified
	selectedPort := *gnssPort
	if selectedPort == "" {
		ports, err := listPorts()
		if err != nil {
			consoleLogger.Fatalf("Error listing serial ports: %v", err)
		}
		if len(ports) == 0 {
			consoleLogger.Fatal("No serial ports found. Please check your connections.")
		}
		selectedPort = selectPort(ports)
	}

	// Create the RTK application
	app := &RTKApp{
		logger:      logger,
		rtcmFilters: defaultRTCMFilters(),
		stopChan:    make(chan struct{}),
		status: RTKStatus{
			Status: rtkStatusNone,
			Time:   time.Now(),
		},
	}

	// Connect to the NTRIP server
	consoleLogger.Printf("Connecting to NTRIP server %s:%s...\n", *ntripServer, *ntripPort)
	app.ntripClient, err = ntrip.NewClient(*ntripServer, *ntripPort, *ntripUser, *ntripPassword, *ntripMountpoint)
	if err != nil {
		consoleLogger.Fatalf("Failed to create NTRIP client: %v", err)
	}

	err = app.ntripClient.Connect()
	if err != nil {
		consoleLogger.Fatalf("Failed to connect to NTRIP server: %v", err)
	}
	defer app.ntripClient.Disconnect()
	consoleLogger.Println("Connected to NTRIP server successfully.")

	// Connect to the GNSS receiver
	consoleLogger.Printf("Connecting to GNSS receiver on port %s...\n", selectedPort)
	app.gnssReceiver, err = ntrip.NewGNSSReceiver(selectedPort)
	if err != nil {
		consoleLogger.Fatalf("Failed to connect to GNSS receiver: %v", err)
	}
	defer app.gnssReceiver.Close()
	consoleLogger.Println("Connected to GNSS receiver successfully.")

	// Start the RTK processor
	consoleLogger.Println("Starting RTK processor...")
	app.rtkProcessor, err = ntrip.NewRTKProcessor(app.gnssReceiver, app.ntripClient)
	if err != nil {
		consoleLogger.Fatalf("Failed to create RTK processor: %v", err)
	}

	err = app.rtkProcessor.Start()
	if err != nil {
		consoleLogger.Fatalf("Failed to start RTK processing: %v", err)
	}
	defer app.rtkProcessor.Stop()
	consoleLogger.Println("RTK processor started successfully.")

	// Start the GNSS data reader
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		buffer := make([]byte, 1024)

		for {
			select {
			case <-app.stopChan:
				return
			default:
				n, err := app.gnssReceiver.Read(buffer)
				if err != nil {
					app.logger.Printf("Error reading from GNSS receiver: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Process NMEA data
					app.updateStatusFromNMEA(buffer[:n])
				}
			}
		}
	}()

	// Start the RTCM data reader
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		buffer := make([]byte, 2048) // RTCM messages can be larger

		for {
			select {
			case <-app.stopChan:
				return
			default:
				n, err := app.ntripClient.Read(buffer)
				if err != nil {
					app.logger.Printf("Error reading from NTRIP server: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Process RTCM data
					app.processRTCM(buffer[:n])
				}
			}
		}
	}()

	// Start the status display
	app.wg.Add(1)
	go app.displayStatus(consoleLogger)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop all goroutines
	close(app.stopChan)
	app.wg.Wait()

	consoleLogger.Println("Application shutdown complete.")
}

// defaultRTCMFilters returns the default RTCM filters
func defaultRTCMFilters() []RTCMFilter {
	return []RTCMFilter{
		// Filter out unwanted message types
		func(msg rtcm.Message) bool {
			// Filter out MT 4094 (Trimble proprietary)
			if msg.Number() == 4094 {
				return false
			}
			// Filter out MT 1013 (System parameters)
			if msg.Number() == 1013 {
				return false
			}
			return true
		},
	}
}

// displayStatus displays the RTK status in a CLI dashboard
func (app *RTKApp) displayStatus(logger *log.Logger) {
	defer app.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			app.statusMutex.Lock()
			status := app.status
			app.statusMutex.Unlock()

			// Get RTK stats
			stats := app.rtkProcessor.GetStats()

			// Clear screen
			fmt.Print("\033[H\033[2J")

			// Display header
			fmt.Println("GNSS RTK Position Calculator")
			fmt.Println("---------------------------")
			fmt.Printf("Time: %s\n\n", status.Time.Format("2006-01-02 15:04:05"))

			// Display position
			fmt.Println("Position Information:")
			fmt.Printf("  Latitude:  %.6f°\n", status.Latitude)
			fmt.Printf("  Longitude: %.6f°\n", status.Longitude)
			fmt.Printf("  Altitude:  %.2f m\n", status.Altitude)
			fmt.Printf("  Accuracy:  %.2f m\n", status.Accuracy)

			// Display RTK status
			fmt.Println("\nRTK Status:")
			statusStr := "UNKNOWN"
			switch status.Status {
			case rtkStatusNone:
				statusStr = "NONE"
			case rtkStatusSingle:
				statusStr = "SINGLE"
			case rtkStatusFloat:
				statusStr = "FLOAT"
			case rtkStatusFix:
				statusStr = "FIX"
			}
			fmt.Printf("  Status:    %s\n", statusStr)
			fmt.Printf("  Satellites: %d\n", status.NSats)
			fmt.Printf("  HDOP:      %.2f\n", status.HDOP)
			fmt.Printf("  Age:       %.1f s\n", status.Age)

			// Display statistics
			fmt.Println("\nRTK Statistics:")
			fmt.Printf("  Rover Observations: %d\n", stats.RoverObs)
			fmt.Printf("  Base Observations:  %d\n", stats.BaseObs)
			fmt.Printf("  Solutions:          %d\n", stats.Solutions)
			fmt.Printf("  Fix Ratio:          %.2f%%\n", stats.FixRatio*100)

			// Display help
			fmt.Println("\nPress Ctrl+C to exit")
		}
	}
}

// listPorts lists all available serial ports with details
func listPorts() ([]string, error) {
	// Get detailed port list
	portDetails, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	var portNames []string
	fmt.Println("Available serial ports:")
	for i, port := range portDetails {
		portInfo := fmt.Sprintf("%d: %s", i+1, port.Name)
		if port.IsUSB {
			portInfo += fmt.Sprintf(" [USB: VID:%04X PID:%04X %s]",
				port.VID, port.PID, port.Product)
		}
		fmt.Println(portInfo)
		portNames = append(portNames, port.Name)
	}

	return portNames, nil
}

// selectPort prompts the user to select a port from the list
func selectPort(ports []string) string {
	if len(ports) == 1 {
		fmt.Printf("Only one port available. Using %s\n", ports[0])
		return ports[0]
	}

	fmt.Println("Please select a port by number:")
	for i, port := range ports {
		fmt.Printf("%d: %s\n", i+1, port)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter port number: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var selection int
		_, err := fmt.Sscanf(input, "%d", &selection)
		if err == nil && selection > 0 && selection <= len(ports) {
			return ports[selection-1]
		}
		fmt.Println("Invalid selection. Please try again.")
	}
}

// updateStatusFromNMEA updates the RTK status from NMEA data
func (app *RTKApp) updateStatusFromNMEA(data []byte) {
	// Convert to string and split into lines
	dataStr := string(data)
	lines := strings.Split(dataStr, "\r\n")

	for _, line := range lines {
		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse NMEA sentence
		sentence, err := nmea.Parse(line)
		if err != nil {
			continue
		}

		// Process GGA sentences for position and RTK status
		if sentence.DataType() == nmea.TypeGGA {
			gga := sentence.(nmea.GGA)

			app.statusMutex.Lock()

			// Update position
			app.status.Latitude = gga.Latitude
			app.status.Longitude = gga.Longitude
			app.status.Altitude = gga.Altitude
			app.status.NSats = gga.NumSatellites
			app.status.HDOP = gga.HDOP
			app.status.Age = gga.DGPSAge
			app.status.Time = time.Now()

			// Update RTK status based on quality indicator
			switch gga.FixQuality {
			case 0:
				app.status.Status = rtkStatusNone
			case 1:
				app.status.Status = rtkStatusSingle
			case 2:
				app.status.Status = rtkStatusFloat
			case 4, 5:
				app.status.Status = rtkStatusFix
			}

			// Update accuracy based on HDOP
			app.status.Accuracy = gga.HDOP * 2.5 // Rough estimate

			app.statusMutex.Unlock()

			// Log the position update
			app.logger.Printf("Position update: Lat=%.6f, Lon=%.6f, Alt=%.2f, Status=%d, Sats=%d",
				gga.Latitude, gga.Longitude, gga.Altitude, gga.FixQuality, gga.NumSatellites)
		}
	}
}

// processRTCM processes RTCM data
func (app *RTKApp) processRTCM(data []byte) {
	// Parse RTCM messages
	messages, err := rtcm.ParseMessages(data)
	if err != nil {
		app.logger.Printf("Error parsing RTCM messages: %v", err)
		return
	}

	// Apply filters
	var filteredMessages []rtcm.Message
	for _, msg := range messages {
		keep := true
		for _, filter := range app.rtcmFilters {
			if !filter(msg) {
				keep = false
				break
			}
		}
		if keep {
			filteredMessages = append(filteredMessages, msg)
		} else {
			app.logger.Printf("Filtered out RTCM message type %d", msg.Number())
		}
	}

	// Log the message types
	for _, msg := range filteredMessages {
		app.logger.Printf("Received RTCM message type %d, length %d bytes",
			msg.Number(), len(msg.Serialize()))
	}
}
