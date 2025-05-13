package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bramburn/go_ntrip/internal/device"
	"github.com/bramburn/go_ntrip/internal/ntrip"
	"github.com/bramburn/go_ntrip/internal/parser"
	"github.com/bramburn/go_ntrip/internal/position"
	"github.com/bramburn/go_ntrip/internal/rtk"
)

// CLI represents the command-line interface
type CLI struct {
	device  device.GNSSDevice
	reader  *bufio.Reader
	running bool
}

// NewCLI creates a new CLI
func NewCLI(device device.GNSSDevice) *CLI {
	return &CLI{
		device:  device,
		reader:  bufio.NewReader(os.Stdin),
		running: false,
	}
}

// Start starts the CLI
func (c *CLI) Start() {
	c.running = true
	c.showWelcome()
	c.mainLoop()
}

// Stop stops the CLI
func (c *CLI) Stop() {
	c.running = false
}

// showWelcome displays the welcome message
func (c *CLI) showWelcome() {
	fmt.Println("\nTOPGNSS TOP708 GNSS Receiver Communication")
	fmt.Println("------------------------------------------")
	c.showHelp()
}

// showHelp displays the help message
func (c *CLI) showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  monitor       - Continuously display raw data")
	fmt.Println("  nmea          - Monitor and parse NMEA sentences")
	fmt.Println("  rtcm          - Monitor RTCM3.3 messages")
	fmt.Println("  ubx           - Monitor UBX protocol messages")
	fmt.Println("  baudrate <n>  - Change baud rate (e.g., baudrate 115200)")
	fmt.Println("  ntrip-pos     - Connect to NTRIP server and get fixed position")
	fmt.Println("  ntrip-avg     - Connect to NTRIP server and average position samples")
	fmt.Println("  help          - Show this help message")
	fmt.Println("  exit          - Quit the application")
	fmt.Println("Or type any command to send directly to the receiver")
}

// mainLoop runs the main command loop
func (c *CLI) mainLoop() {
	for c.running {
		fmt.Print("\n> ")
		command, _ := c.reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if command == "exit" {
			fmt.Println("Exiting...")
			c.running = false
			return
		}

		c.handleCommand(command)
	}
}

// handleCommand processes a user command
func (c *CLI) handleCommand(command string) {
	switch {
	case command == "help":
		c.showHelp()

	case command == "monitor":
		fmt.Println("Monitoring raw device output. Press Enter to stop.")
		c.monitorRawData()

	case command == "nmea":
		fmt.Println("Monitoring and parsing NMEA sentences. Press Enter to stop.")
		c.monitorNMEA()

	case command == "rtcm":
		fmt.Println("Monitoring RTCM3.3 messages. Press Enter to stop.")
		c.monitorRTCM()

	case command == "ubx":
		fmt.Println("Monitoring UBX protocol messages. Press Enter to stop.")
		c.monitorUBX()

	case command == "ntrip-pos":
		fmt.Println("Connecting to NTRIP server to get fixed position...")
		c.getNtripPosition()

	case command == "ntrip-avg":
		fmt.Println("Connecting to NTRIP server to average position samples...")
		c.getAveragedPosition()

	case strings.HasPrefix(command, "baudrate "):
		c.changeBaudRate(command)

	case command != "":
		c.sendCommand(command)
	}
}

// monitorRawData displays raw data from the device
func (c *CLI) monitorRawData() {
	if !c.device.IsConnected() {
		fmt.Println("Device not connected.")
		return
	}

	// Create a channel to signal stopping
	stopChan := make(chan bool)
	buffer := make([]byte, 1024)

	// Start goroutine to read from device
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := c.device.ReadRaw(buffer)
				if err != nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if n > 0 {
					fmt.Print(string(buffer[:n]))
				}
			}
		}
	}()

	// Wait for Enter key to stop
	c.reader.ReadString('\n')
	stopChan <- true
	fmt.Println("\nStopped monitoring.")
}

// monitorNMEA monitors and parses NMEA sentences
func (c *CLI) monitorNMEA() {
	if !c.device.IsConnected() {
		fmt.Println("Device not connected.")
		return
	}

	// Create a handler for NMEA data
	handler := &NMEAHandler{}

	// Create monitoring config
	config := device.DefaultMonitorConfig(device.ProtocolNMEA, handler)

	// Start monitoring
	if d, ok := c.device.(*device.TOPGNSSDevice); ok {
		err := d.MonitorNMEA(config)
		if err != nil {
			fmt.Printf("Error starting NMEA monitoring: %v\n", err)
			return
		}
	} else {
		fmt.Println("Device does not support NMEA monitoring.")
		return
	}

	// Wait for Enter key to stop
	c.reader.ReadString('\n')

	// Stop monitoring
	if d, ok := c.device.(*device.TOPGNSSDevice); ok {
		d.StopMonitoring()
	}

	fmt.Println("\nStopped monitoring NMEA data.")
}

// monitorRTCM monitors RTCM messages
func (c *CLI) monitorRTCM() {
	fmt.Println("RTCM monitoring not implemented yet.")
	c.reader.ReadString('\n')
}

// monitorUBX monitors UBX messages
func (c *CLI) monitorUBX() {
	fmt.Println("UBX monitoring not implemented yet.")
	c.reader.ReadString('\n')
}

// changeBaudRate changes the baud rate
func (c *CLI) changeBaudRate(command string) {
	parts := strings.Split(command, " ")
	if len(parts) != 2 {
		fmt.Println("Invalid baudrate command. Usage: baudrate <rate>")
		return
	}

	var newBaudRate int
	_, err := fmt.Sscanf(parts[1], "%d", &newBaudRate)
	if err != nil {
		fmt.Printf("Invalid baud rate: %s\n", parts[1])
		return
	}

	err = c.device.ChangeBaudRate(newBaudRate)
	if err != nil {
		fmt.Printf("Error changing baud rate: %v\n", err)
		return
	}

	fmt.Printf("Baud rate changed to %d successfully.\n", newBaudRate)
}

// sendCommand sends a command to the device
func (c *CLI) sendCommand(command string) {
	if !c.device.IsConnected() {
		fmt.Println("Device not connected.")
		return
	}

	err := c.device.WriteCommand(command)
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		return
	}

	// Read response
	buffer := make([]byte, 1024)
	time.Sleep(500 * time.Millisecond) // Give device time to respond

	n, err := c.device.ReadRaw(buffer)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	if n > 0 {
		fmt.Print("Response: ")
		fmt.Println(string(buffer[:n]))
	} else {
		fmt.Println("No response received.")
	}
}

// NMEAHandler implements device.DataHandler for NMEA data
type NMEAHandler struct {
	parser *parser.NMEAParser
}

// NewNMEAHandler creates a new NMEA handler
func NewNMEAHandler() *NMEAHandler {
	return &NMEAHandler{
		parser: parser.NewNMEAParser(),
	}
}

// HandleNMEA handles NMEA sentences
func (h *NMEAHandler) HandleNMEA(sentence parser.NMEASentence) {
	// Format based on sentence type
	switch {
	case strings.HasSuffix(sentence.Type, "GGA"):
		fmt.Printf("\n[%s] Global Positioning System Fix Data\n", sentence.Type)
		if len(sentence.Fields) >= 14 {
			fmt.Printf("  Time: %s UTC\n", h.parser.FormatTime(sentence.Fields[0]))
			fmt.Printf("  Latitude: %s %s\n", h.parser.FormatLatLon(sentence.Fields[1]), sentence.Fields[2])
			fmt.Printf("  Longitude: %s %s\n", h.parser.FormatLatLon(sentence.Fields[3]), sentence.Fields[4])
			fmt.Printf("  Fix Quality: %s\n", h.parser.GetFixQuality(sentence.Fields[5]))
			fmt.Printf("  Satellites: %s\n", sentence.Fields[6])
			fmt.Printf("  HDOP: %s\n", sentence.Fields[7])
			fmt.Printf("  Altitude: %s meters\n", sentence.Fields[8])
			fmt.Printf("  Geoid Height: %s meters\n", sentence.Fields[10])
		}
	// Add other sentence types as needed
	default:
		fmt.Printf("\n[%s] Raw NMEA Sentence\n", sentence.Type)
		for i, field := range sentence.Fields {
			fmt.Printf("  Field %d: %s\n", i+1, field)
		}
	}
}

// HandleRTCM handles RTCM messages
func (h *NMEAHandler) HandleRTCM(message parser.RTCMMessage) {
	// Not used for NMEA handler
}

// HandleUBX handles UBX messages
func (h *NMEAHandler) HandleUBX(message parser.UBXMessage) {
	// Not used for NMEA handler
}

// getNtripPosition connects to an NTRIP server and gets the fixed position
func (c *CLI) getNtripPosition() {
	// Prompt for NTRIP server details
	fmt.Println("Enter NTRIP server details:")
	fmt.Print("Server address (e.g., 192.168.0.64): ")
	address, _ := c.reader.ReadString('\n')
	address = strings.TrimSpace(address)

	fmt.Print("Port (e.g., 2101): ")
	portStr, _ := c.reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)

	fmt.Print("Username: ")
	username, _ := c.reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Mountpoint: ")
	mountpoint, _ := c.reader.ReadString('\n')
	mountpoint = strings.TrimSpace(mountpoint)

	// Construct URL
	url := fmt.Sprintf("http://%s:%s", address, portStr)

	// Create NTRIP client
	client := ntrip.NewClient(url, username, password, mountpoint)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect to NTRIP server
	fmt.Printf("Connecting to NTRIP server at %s...\n", url)
	stream, err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error connecting to NTRIP server: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Println("Connected to NTRIP server. Receiving RTCM data...")

	// Create a channel to signal stopping
	stopChan := make(chan bool)

	// Create a channel for position data
	positionChan := make(chan *position.Position)

	// Create RTK processor
	processor := rtk.NewProcessor()

	// Start processing
	processor.StartProcessing()

	// Start goroutine to collect solutions
	go func() {
		solutionChan := processor.GetSolutionChannel()

		for {
			select {
			case <-stopChan:
				return
			case solution := <-solutionChan:
				// Only process fixed solutions
				if solution.Status >= rtk.StatusFix {
					// Convert solution to position
					pos := solution.ToPosition()

					// Send position to channel
					positionChan <- pos
				} else {
					// Display current fix quality
					fmt.Printf("Current fix quality: %s\r",
						position.GetFixQualityDescription(solution.Status))
				}
			}
		}
	}()

	// Start goroutine to read from NTRIP stream and process RTCM data
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := stream.Read(buffer)
				if err != nil {
					fmt.Printf("Error reading from NTRIP stream: %v\n", err)
					return
				}
				if n > 0 {
					// Process RTCM data
					processor.ProcessRTCM(buffer[:n])
				}
			}
		}
	}()

	fmt.Println("Waiting for RTK fixed position...")
	fmt.Println("Press Enter to stop.")

	// Wait for position or user input
	go func() {
		c.reader.ReadString('\n')
		stopChan <- true
	}()

	// Wait for position data
	select {
	case pos := <-positionChan:
		// Stop RTK processing
		processor.StopProcessing()

		// Display position information
		fmt.Println("\nReceived fixed position:")
		fmt.Printf("  Latitude: %.8f\n", pos.Latitude)
		fmt.Printf("  Longitude: %.8f\n", pos.Longitude)
		fmt.Printf("  Altitude: %.2f meters\n", pos.Altitude)
		fmt.Printf("  Fix Quality: %s\n", pos.Description)
		fmt.Printf("  Satellites: %d\n", pos.Satellites)
		fmt.Printf("  HDOP: %.2f\n", pos.HDOP)
		fmt.Printf("  Timestamp: %s\n", pos.Timestamp.Format(time.RFC3339))

		// Save position to file
		execPath, err := os.Executable()
		if err != nil {
			execPath = "."
		}
		filePath := filepath.Join(filepath.Dir(execPath), "base_position.json")
		err = pos.SaveToFile(filePath)
		if err != nil {
			fmt.Printf("Error saving position to file: %v\n", err)
		} else {
			fmt.Printf("Position saved to %s\n", filePath)
		}

	case <-stopChan:
		// Stop monitoring
		if d, ok := c.device.(*device.TOPGNSSDevice); ok {
			d.StopMonitoring()
		}
		fmt.Println("\nStopped NTRIP connection.")
	}
}

// NtripPositionHandler implements device.DataHandler for NMEA data with position extraction
type NtripPositionHandler struct {
	parser       *parser.NMEAParser
	positionChan chan<- *position.Position
}

// HandleNMEA handles NMEA sentences and extracts position information
func (h *NtripPositionHandler) HandleNMEA(sentence parser.NMEASentence) {
	// Only process GGA sentences
	if strings.HasSuffix(sentence.Type, "GGA") && len(sentence.Fields) >= 14 {
		// Extract fix quality
		fixQuality := 0
		if sentence.Fields[5] != "" {
			fixQuality, _ = strconv.Atoi(sentence.Fields[5])
		}

		// Only process RTK fixed positions (fix quality 4 or 5)
		if fixQuality >= 4 {
			// Extract position
			pos, err := position.ExtractFromGGA(sentence)
			if err == nil {
				// Send position to channel
				h.positionChan <- pos
			}
		} else {
			// Display current fix quality
			fmt.Printf("Current fix quality: %s\r", h.parser.GetFixQuality(sentence.Fields[5]))
		}
	}
}

// HandleRTCM handles RTCM messages
func (h *NtripPositionHandler) HandleRTCM(message parser.RTCMMessage) {
	// Not used for NMEA handler
}

// getAveragedPosition connects to an NTRIP server and averages position samples
func (c *CLI) getAveragedPosition() {
	// Prompt for NTRIP server details
	fmt.Println("Enter NTRIP server details:")
	fmt.Print("Server address (e.g., 192.168.0.64): ")
	address, _ := c.reader.ReadString('\n')
	address = strings.TrimSpace(address)

	fmt.Print("Port (e.g., 2101): ")
	portStr, _ := c.reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)

	fmt.Print("Username: ")
	username, _ := c.reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Mountpoint: ")
	mountpoint, _ := c.reader.ReadString('\n')
	mountpoint = strings.TrimSpace(mountpoint)

	fmt.Print("Minimum fix quality (4=RTK Fixed, 5=Float RTK): ")
	minFixQualityStr, _ := c.reader.ReadString('\n')
	minFixQualityStr = strings.TrimSpace(minFixQualityStr)
	minFixQuality, err := strconv.Atoi(minFixQualityStr)
	if err != nil || minFixQuality < 0 || minFixQuality > 8 {
		fmt.Println("Invalid fix quality. Using default (4 - RTK Fixed).")
		minFixQuality = 4
	}

	fmt.Print("Number of samples to collect (default: 60): ")
	sampleCountStr, _ := c.reader.ReadString('\n')
	sampleCountStr = strings.TrimSpace(sampleCountStr)
	sampleCount, err := strconv.Atoi(sampleCountStr)
	if err != nil || sampleCount <= 0 {
		fmt.Println("Invalid sample count. Using default (60).")
		sampleCount = 60
	}

	fmt.Print("Output file path (default: ./base_position_avg.json): ")
	filePath, _ := c.reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		execPath, err := os.Executable()
		if err != nil {
			execPath = "."
		}
		filePath = filepath.Join(filepath.Dir(execPath), "base_position_avg.json")
	}

	// Construct URL
	url := fmt.Sprintf("http://%s:%s", address, portStr)

	// Create NTRIP client
	client := ntrip.NewClient(url, username, password, mountpoint)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Connect to NTRIP server
	fmt.Printf("Connecting to NTRIP server at %s...\n", url)
	stream, err := client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error connecting to NTRIP server: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Println("Connected to NTRIP server. Receiving RTCM data...")

	// Create a channel to signal stopping
	stopChan := make(chan bool)

	// Create a channel for position data
	positionChan := make(chan *position.Position)

	// Create position averager
	averager := position.NewPositionAverager(minFixQuality)

	// Create RTK processor
	processor := rtk.NewProcessor()

	// Start processing
	processor.StartProcessing()

	// Start goroutine to collect solutions
	go func() {
		solutionChan := processor.GetSolutionChannel()
		currentCount := 0

		for {
			select {
			case <-stopChan:
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
						currentCount, sampleCount, position.GetFixQualityDescription(pos.FixQuality))

					// Check if we've collected enough samples
					if currentCount >= sampleCount {
						// Get averaged position
						pos, stats, err := averager.GetAveragedPosition()
						if err == nil {
							// Attach stats to position
							pos.Stats = stats

							// Send position to channel
							positionChan <- pos
						}
					}
				} else {
					// Display current fix quality if sample was rejected
					fmt.Printf("Current fix quality: %s (not used)\r",
						position.GetFixQualityDescription(pos.FixQuality))
				}
			}
		}
	}()

	// Start goroutine to read from NTRIP stream and process RTCM data
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := stream.Read(buffer)
				if err != nil {
					fmt.Printf("Error reading from NTRIP stream: %v\n", err)
					return
				}
				if n > 0 {
					// Process RTCM data
					processor.ProcessRTCM(buffer[:n])
				}
			}
		}
	}()

	fmt.Printf("Collecting position samples (minimum fix quality: %s)...\n",
		position.GetFixQualityDescription(minFixQuality))
	fmt.Printf("Will collect %d samples. Press Enter to stop early.\n", sampleCount)

	// Wait for user input to stop early
	go func() {
		c.reader.ReadString('\n')
		stopChan <- true
	}()

	// Wait for position data or completion
	select {
	case pos := <-positionChan:
		// Stop RTK processing
		processor.StopProcessing()

		// Get stats
		stats := pos.Stats

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
		err = pos.SaveToFile(filePath)
		if err != nil {
			fmt.Printf("Error saving position to file: %v\n", err)
		} else {
			fmt.Printf("Position saved to %s\n", filePath)
		}

	case <-stopChan:
		// Stop RTK processing
		processor.StopProcessing()
		fmt.Println("\nStopped NTRIP connection.")

		// Check if we have any samples
		if averager.GetSampleCount() > 0 {
			// Get averaged position
			pos, stats, err := averager.GetAveragedPosition()
			if err != nil {
				fmt.Printf("Error getting averaged position: %v\n", err)
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
			err = position.SavePositionWithStats(pos, stats, filePath)
			if err != nil {
				fmt.Printf("Error saving position to file: %v\n", err)
			} else {
				fmt.Printf("Position saved to %s\n", filePath)
			}
		} else {
			fmt.Println("No position samples collected.")
		}
	}
}

// PositionAveragerHandler implements device.DataHandler for NMEA data with position averaging
type PositionAveragerHandler struct {
	parser       *parser.NMEAParser
	averager     *position.PositionAverager
	positionChan chan<- *position.Position
	sampleCount  int
	currentCount int
}

// HandleNMEA handles NMEA sentences and extracts position information for averaging
func (h *PositionAveragerHandler) HandleNMEA(sentence parser.NMEASentence) {
	// Only process GGA sentences
	if strings.HasSuffix(sentence.Type, "GGA") && len(sentence.Fields) >= 14 {
		// Extract fix quality
		fixQuality := 0
		if sentence.Fields[5] != "" {
			fixQuality, _ = strconv.Atoi(sentence.Fields[5])
		}

		// Extract position
		pos, err := position.ExtractFromGGA(sentence)
		if err == nil {
			// Create sample
			sample := position.PositionSample{
				Latitude:   pos.Latitude,
				Longitude:  pos.Longitude,
				Altitude:   pos.Altitude,
				FixQuality: fixQuality,
				Timestamp:  pos.Timestamp,
			}

			// Add sample to averager
			if h.averager.AddSample(sample) {
				// Increment count if sample was accepted
				h.currentCount++
				fmt.Printf("Sample %d/%d collected (Fix: %s)\r",
					h.currentCount, h.sampleCount, position.GetFixQualityDescription(fixQuality))

				// Check if we've collected enough samples
				if h.currentCount >= h.sampleCount {
					// Get averaged position
					pos, stats, err := h.averager.GetAveragedPosition()
					if err == nil {
						// Attach stats to position
						pos.Stats = stats

						// Send position to channel
						h.positionChan <- pos
					}
				}
			} else {
				// Display current fix quality if sample was rejected
				fmt.Printf("Current fix quality: %s (not used)\r",
					position.GetFixQualityDescription(fixQuality))
			}
		}
	}
}

// HandleUBX handles UBX messages
func (h *PositionAveragerHandler) HandleUBX(message parser.UBXMessage) {
	// Not used for NMEA handler
}

// HandleRTCM handles RTCM messages
func (h *PositionAveragerHandler) HandleRTCM(message parser.RTCMMessage) {
	// Not used for NMEA handler
}

// HandleUBX handles UBX message
// HandleUBX handles UBX messages
func (h *NtripPositionHandler) HandleUBX(message parser.UBXMessage) {
	// Not used for NMEA handler
}
