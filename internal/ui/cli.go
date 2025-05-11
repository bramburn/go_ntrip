package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bramburn/go_ntrip/internal/device"
	"github.com/bramburn/go_ntrip/internal/parser"
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
		device: device,
		reader: bufio.NewReader(os.Stdin),
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
