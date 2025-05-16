package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-gnss/rtcm"
	"go.bug.st/serial"
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

// NMEA sentence structure
type NMEASentence struct {
	Type     string
	Fields   []string
	Valid    bool
	Checksum string
}

// UBX message structure
type UBXMessage struct {
	Class   byte
	ID      byte
	Length  uint16
	Payload []byte
	Valid   bool
}

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

func main() {
	// List available ports with details
	ports, err := listPorts()
	if err != nil {
		log.Fatalf("Error listing serial ports: %v", err)
	}

	if len(ports) == 0 {
		log.Fatal("No serial ports found. Please check your connections.")
	}

	// Select port to use - always prompt for selection
	portName := selectPort(ports)

	// Configure serial port for TOPGNSS TOP708
	mode := &serial.Mode{
		BaudRate: defaultBaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	// Open the port with error handling
	fmt.Printf("Opening port %s with baud rate %d...\n", portName, defaultBaudRate)
	port, err := serial.Open(portName, mode)
	if err != nil {
		handleConnectionError(err, portName)
		return
	}
	defer port.Close()

	// Set read timeout
	err = port.SetReadTimeout(readTimeout)
	if err != nil {
		log.Printf("Warning: Error setting read timeout: %v", err)
		fmt.Println("Continuing with default timeout settings...")
	}

	fmt.Println("Port opened successfully. Waiting for device to initialize...")
	time.Sleep(2 * time.Second) // Give the device time to initialize

	// Verify connection by checking for GNSS data
	if !verifyConnection(port) {
		fmt.Println("Unable to verify GNSS data. The device may not be sending data.")
		fmt.Println("Do you want to continue anyway? (y/n)")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Exiting...")
			return
		}
	}

	// Start interactive session
	interactWithDevice(port)
}

// verifyConnection checks if the device is sending valid GNSS data
func verifyConnection(port serial.Port) bool {
	fmt.Println("Checking for GNSS data...")
	buffer := make([]byte, 1024)

	// Try to read data for up to 5 seconds
	for i := 0; i < 10; i++ {
		n, err := port.Read(buffer)
		if err != nil {
			log.Printf("Error reading from port: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if n > 0 {
			data := string(buffer[:n])
			// Check for NMEA sentences
			if strings.Contains(data, "$GN") || strings.Contains(data, "$GP") {
				fmt.Println("GNSS data detected!")
				return true
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

// handleConnectionError provides detailed error handling for connection issues
func handleConnectionError(err error, portName string) {
	log.Printf("Error opening serial port %s: %v", portName, err)

	fmt.Println("\nTroubleshooting tips:")
	fmt.Println("1. Check if the GNSS receiver is properly connected")
	fmt.Println("2. Verify that no other application is using the port")
	fmt.Println("3. Try a different USB port")
	fmt.Println("4. Check if the correct drivers are installed")
	fmt.Println("5. Try restarting the GNSS receiver")

	// Check for specific error types
	errStr := err.Error()
	if strings.Contains(errStr, "access denied") || strings.Contains(errStr, "permission denied") {
		fmt.Println("\nPermission issue detected:")
		fmt.Println("- Try running the application with administrator privileges")
		fmt.Println("- Check if another application is using the port")
	} else if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such file") {
		fmt.Println("\nPort not found:")
		fmt.Println("- The selected port may no longer be available")
		fmt.Println("- Try reconnecting the device and restarting the application")
	} else if strings.Contains(errStr, "timeout") {
		fmt.Println("\nConnection timeout:")
		fmt.Println("- The device is not responding")
		fmt.Println("- Check if the baud rate (38400) matches your device configuration")
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

// interactWithDevice handles the interactive session with the GNSS device
func interactWithDevice(port serial.Port) {
	reader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, 2048) // Larger buffer for RTCM data

	fmt.Println("\nTOPGNSS TOP708 GNSS Receiver Communication")
	fmt.Println("------------------------------------------")
	fmt.Println("Available commands:")
	fmt.Println("  monitor       - Continuously display raw data")
	fmt.Println("  nmea          - Monitor and parse NMEA sentences")
	fmt.Println("  rtcm          - Monitor RTCM3.3 messages")
	fmt.Println("  ubx           - Monitor UBX protocol messages")
	fmt.Println("  baudrate <n>  - Change baud rate (e.g., baudrate 115200)")
	fmt.Println("  help          - Show this help message")
	fmt.Println("  exit          - Quit the application")
	fmt.Println("Or type any command to send directly to the receiver")

	for {
		fmt.Print("\n> ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if command == "exit" {
			fmt.Println("Exiting...")
			return
		}

		if command == "help" {
			fmt.Println("Available commands:")
			fmt.Println("  monitor       - Continuously display raw data")
			fmt.Println("  nmea          - Monitor and parse NMEA sentences")
			fmt.Println("  rtcm          - Monitor RTCM3.3 messages")
			fmt.Println("  ubx           - Monitor UBX protocol messages")
			fmt.Println("  baudrate <n>  - Change baud rate (e.g., baudrate 115200)")
			fmt.Println("  help          - Show this help message")
			fmt.Println("  exit          - Quit the application")
			continue
		}

		if command == "monitor" {
			fmt.Println("Monitoring raw device output. Press Enter to stop.")
			monitorRawData(port, reader)
			continue
		}

		if command == "nmea" {
			fmt.Println("Monitoring and parsing NMEA sentences. Press Enter to stop.")
			monitorNMEA(port, reader)
			continue
		}

		if command == "rtcm" {
			fmt.Println("Monitoring RTCM3.3 messages. Press Enter to stop.")
			monitorRTCM(port, reader)
			continue
		}

		if command == "ubx" {
			fmt.Println("Monitoring UBX protocol messages. Press Enter to stop.")
			monitorUBX(port, reader)
			continue
		}

		if strings.HasPrefix(command, "baudrate ") {
			parts := strings.Split(command, " ")
			if len(parts) != 2 {
				fmt.Println("Invalid baudrate command. Usage: baudrate <rate>")
				continue
			}

			var newBaudRate int
			_, err := fmt.Sscanf(parts[1], "%d", &newBaudRate)
			if err != nil {
				fmt.Printf("Invalid baud rate: %s\n", parts[1])
				continue
			}

			changeBaudRate(port, newBaudRate)
			continue
		}

		// Send command to device
		if command != "" {
			// Add newline if not present
			if !strings.HasSuffix(command, "\r\n") {
				command += "\r\n"
			}

			_, err := port.Write([]byte(command))
			if err != nil {
				log.Printf("Error writing to port: %v", err)
				continue
			}

			// Read response
			time.Sleep(500 * time.Millisecond) // Give device time to respond
			n, err := port.Read(buffer)
			if err != nil {
				log.Printf("Error reading from port: %v", err)
				continue
			}

			if n > 0 {
				fmt.Print("Response: ")
				fmt.Println(string(buffer[:n]))
			} else {
				fmt.Println("No response received.")
			}
		}
	}
}

// monitorRawData displays raw data from the device
func monitorRawData(port serial.Port, reader *bufio.Reader) {
	buffer := make([]byte, 1024)
	stopChan := make(chan bool)

	// Start goroutine to read from port
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := port.Read(buffer)
				if err != nil {
					log.Printf("Error reading from port: %v", err)
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
	reader.ReadString('\n')
	stopChan <- true
	fmt.Println("\nStopped monitoring.")
}

// monitorNMEA monitors and parses NMEA sentences
func monitorNMEA(port serial.Port, reader *bufio.Reader) {
	buffer := make([]byte, 1024)
	stopChan := make(chan bool)
	dataBuffer := ""

	// Start goroutine to read and parse NMEA data
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := port.Read(buffer)
				if err != nil {
					log.Printf("Error reading from port: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Add new data to buffer
					dataBuffer += string(buffer[:n])

					// Process complete NMEA sentences
					for {
						// Find start and end of NMEA sentence
						startIdx := strings.Index(dataBuffer, "$")
						if startIdx == -1 {
							break
						}

						endIdx := strings.Index(dataBuffer[startIdx:], "\r\n")
						if endIdx == -1 {
							break
						}
						endIdx += startIdx

						// Extract and parse the sentence
						sentence := dataBuffer[startIdx:endIdx]
						parsedSentence := parseNMEA(sentence)

						// Display parsed data
						if parsedSentence.Valid {
							displayNMEA(parsedSentence)
						}

						// Remove processed data from buffer
						if endIdx+2 <= len(dataBuffer) {
							dataBuffer = dataBuffer[endIdx+2:]
						} else {
							dataBuffer = ""
						}
					}
				}
			}
		}
	}()

	// Wait for Enter key to stop
	reader.ReadString('\n')
	stopChan <- true
	fmt.Println("\nStopped monitoring NMEA data.")
}

// parseNMEA parses an NMEA sentence
func parseNMEA(sentence string) NMEASentence {
	result := NMEASentence{
		Valid: false,
	}

	// Check for minimum length
	if len(sentence) < 6 {
		return result
	}

	// Check for valid start character
	if sentence[0] != '$' {
		return result
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
		return result
	}

	// Extract sentence type
	result.Type = strings.TrimPrefix(fields[0], "$")
	result.Fields = fields[1:]
	result.Valid = true

	return result
}

// displayNMEA formats and displays NMEA sentence information
func displayNMEA(sentence NMEASentence) {
	// Format based on sentence type
	switch {
	case strings.HasSuffix(sentence.Type, "GGA"):
		fmt.Printf("\n[%s] Global Positioning System Fix Data\n", sentence.Type)
		if len(sentence.Fields) >= 14 {
			fmt.Printf("  Time: %s UTC\n", formatNMEATime(sentence.Fields[0]))
			fmt.Printf("  Latitude: %s %s\n", formatNMEALatLon(sentence.Fields[1]), sentence.Fields[2])
			fmt.Printf("  Longitude: %s %s\n", formatNMEALatLon(sentence.Fields[3]), sentence.Fields[4])
			fmt.Printf("  Fix Quality: %s\n", getFixQuality(sentence.Fields[5]))
			fmt.Printf("  Satellites: %s\n", sentence.Fields[6])
			fmt.Printf("  HDOP: %s\n", sentence.Fields[7])
			fmt.Printf("  Altitude: %s meters\n", sentence.Fields[8])
			fmt.Printf("  Geoid Height: %s meters\n", sentence.Fields[10])
		}

	case strings.HasSuffix(sentence.Type, "RMC"):
		fmt.Printf("\n[%s] Recommended Minimum Navigation Information\n", sentence.Type)
		if len(sentence.Fields) >= 12 {
			fmt.Printf("  Time: %s UTC\n", formatNMEATime(sentence.Fields[0]))

			// Status field
			status := "Void"
			if sentence.Fields[1] == "A" {
				status = "Active"
			}
			fmt.Printf("  Status: %s\n", status)

			fmt.Printf("  Latitude: %s %s\n", formatNMEALatLon(sentence.Fields[2]), sentence.Fields[3])
			fmt.Printf("  Longitude: %s %s\n", formatNMEALatLon(sentence.Fields[4]), sentence.Fields[5])
			fmt.Printf("  Speed: %s knots\n", sentence.Fields[6])
			fmt.Printf("  Course: %s degrees\n", sentence.Fields[7])
			fmt.Printf("  Date: %s\n", formatNMEADate(sentence.Fields[8]))
			fmt.Printf("  Magnetic Variation: %s %s\n", sentence.Fields[9], sentence.Fields[10])
		}

	case strings.HasSuffix(sentence.Type, "GSV"):
		fmt.Printf("\n[%s] Satellites in View\n", sentence.Type)
		if len(sentence.Fields) >= 3 {
			fmt.Printf("  Total Messages: %s\n", sentence.Fields[0])
			fmt.Printf("  Message Number: %s\n", sentence.Fields[1])
			fmt.Printf("  Satellites in View: %s\n", sentence.Fields[2])
		}

	case strings.HasSuffix(sentence.Type, "GSA"):
		fmt.Printf("\n[%s] GPS DOP and Active Satellites\n", sentence.Type)
		if len(sentence.Fields) >= 17 {
			// Mode field
			mode := "Manual"
			if sentence.Fields[0] == "A" {
				mode = "Automatic"
			}
			fmt.Printf("  Mode: %s\n", mode)

			fmt.Printf("  Fix Type: %s\n", getFixType(sentence.Fields[1]))
			fmt.Printf("  PDOP: %s\n", sentence.Fields[14])
			fmt.Printf("  HDOP: %s\n", sentence.Fields[15])
			fmt.Printf("  VDOP: %s\n", sentence.Fields[16])
		}

	case strings.HasSuffix(sentence.Type, "GLL"):
		fmt.Printf("\n[%s] Geographic Position - Latitude/Longitude\n", sentence.Type)
		if len(sentence.Fields) >= 6 {
			fmt.Printf("  Latitude: %s %s\n", formatNMEALatLon(sentence.Fields[0]), sentence.Fields[1])
			fmt.Printf("  Longitude: %s %s\n", formatNMEALatLon(sentence.Fields[2]), sentence.Fields[3])
			fmt.Printf("  Time: %s UTC\n", formatNMEATime(sentence.Fields[4]))

			// Status field
			status := "Invalid"
			if sentence.Fields[5] == "A" {
				status = "Valid"
			}
			fmt.Printf("  Status: %s\n", status)
		}

	default:
		fmt.Printf("\n[%s] Raw NMEA Sentence\n", sentence.Type)
		for i, field := range sentence.Fields {
			fmt.Printf("  Field %d: %s\n", i+1, field)
		}
	}
}

// formatNMEATime formats NMEA time string (HHMMSS.sss)
func formatNMEATime(timeStr string) string {
	if len(timeStr) < 6 {
		return timeStr
	}

	hours := timeStr[0:2]
	minutes := timeStr[2:4]
	seconds := timeStr[4:]

	return hours + ":" + minutes + ":" + seconds
}

// formatNMEADate formats NMEA date string (DDMMYY)
func formatNMEADate(dateStr string) string {
	if len(dateStr) != 6 {
		return dateStr
	}

	day := dateStr[0:2]
	month := dateStr[2:4]
	year := dateStr[4:6]

	return day + "/" + month + "/20" + year
}

// formatNMEALatLon formats latitude/longitude from NMEA format
func formatNMEALatLon(coord string) string {
	if coord == "" {
		return "N/A"
	}
	return coord
}

// getFixQuality returns a description of the fix quality
func getFixQuality(quality string) string {
	switch quality {
	case "0":
		return "Invalid (0)"
	case "1":
		return "GPS Fix (1)"
	case "2":
		return "DGPS Fix (2)"
	case "3":
		return "PPS Fix (3)"
	case "4":
		return "RTK Fix (4)"
	case "5":
		return "Float RTK (5)"
	case "6":
		return "Estimated (6)"
	case "7":
		return "Manual Input (7)"
	case "8":
		return "Simulation (8)"
	default:
		return quality
	}
}

// getFixType returns a description of the fix type
func getFixType(fixType string) string {
	switch fixType {
	case "1":
		return "No Fix (1)"
	case "2":
		return "2D Fix (2)"
	case "3":
		return "3D Fix (3)"
	default:
		return fixType
	}
}

// monitorRTCM monitors RTCM3.3 messages
func monitorRTCM(port serial.Port, reader *bufio.Reader) {
	buffer := make([]byte, 2048) // RTCM messages can be larger
	stopChan := make(chan bool)

	// Start goroutine to read and detect RTCM data
	go func() {
		rtcmBuffer := make([]byte, 0)

		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := port.Read(buffer)
				if err != nil {
					log.Printf("Error reading from port: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Add new data to buffer
					rtcmBuffer = append(rtcmBuffer, buffer[:n]...)

					// Process RTCM messages
					for len(rtcmBuffer) >= 3 {
						// Check for RTCM signature (preamble byte 0xD3 followed by 2 bytes)
						if rtcmBuffer[0] == 0xD3 {
							// Get message length from bytes 2-3 (10 bits)
							if len(rtcmBuffer) < 6 {
								break // Not enough data yet
							}

							// Extract length (10 bits from bytes 1-2)
							length := (int(rtcmBuffer[1]&0x03) << 8) | int(rtcmBuffer[2])
							totalLength := length + 6 // Add header and CRC

							if len(rtcmBuffer) >= totalLength {
								// We have a complete message
								messageType := (int(rtcmBuffer[3]) << 4) | (int(rtcmBuffer[4]) >> 4)
								fmt.Printf("\nRTCM3.3 Message Type: %d, Length: %d bytes\n", messageType, length)

								// Remove processed message from buffer
								rtcmBuffer = rtcmBuffer[totalLength:]
							} else {
								break // Wait for more data
							}
						} else {
							// Not an RTCM message, skip this byte
							rtcmBuffer = rtcmBuffer[1:]
						}
					}
				}
			}
		}
	}()

	// Wait for Enter key to stop
	reader.ReadString('\n')
	stopChan <- true
	fmt.Println("\nStopped monitoring RTCM data.")
}

// monitorUBX monitors UBX protocol messages
func monitorUBX(port serial.Port, reader *bufio.Reader) {
	buffer := make([]byte, 1024)
	stopChan := make(chan bool)

	// Start goroutine to read and detect UBX messages
	go func() {
		ubxBuffer := make([]byte, 0)

		for {
			select {
			case <-stopChan:
				return
			default:
				n, err := port.Read(buffer)
				if err != nil {
					log.Printf("Error reading from port: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Add new data to buffer
					ubxBuffer = append(ubxBuffer, buffer[:n]...)

					// Process UBX messages
					for len(ubxBuffer) >= 6 {
						// Check for UBX signature (0xB5 0x62)
						if ubxBuffer[0] == 0xB5 && ubxBuffer[1] == 0x62 {
							// Get message class and ID
							msgClass := ubxBuffer[2]
							msgID := ubxBuffer[3]

							// Get payload length (little endian)
							payloadLength := int(ubxBuffer[4]) | (int(ubxBuffer[5]) << 8)
							totalLength := payloadLength + 8 // Add header and checksum

							if len(ubxBuffer) >= totalLength {
								// We have a complete message
								fmt.Printf("\nUBX Message: Class=0x%02X, ID=0x%02X, Length=%d bytes\n",
									msgClass, msgID, payloadLength)

								// Display message class information
								displayUBXInfo(msgClass, msgID)

								// Remove processed message from buffer
								ubxBuffer = ubxBuffer[totalLength:]
							} else {
								break // Wait for more data
							}
						} else {
							// Not a UBX message, skip this byte
							ubxBuffer = ubxBuffer[1:]
						}
					}
				}
			}
		}
	}()

	// Wait for Enter key to stop
	reader.ReadString('\n')
	stopChan <- true
	fmt.Println("\nStopped monitoring UBX data.")
}

// displayUBXInfo displays information about UBX message classes and IDs
func displayUBXInfo(msgClass byte, msgID byte) {
	// UBX message class descriptions
	classInfo := ""
	switch msgClass {
	case 0x01:
		classInfo = "NAV (Navigation Results)"
		// NAV message IDs
		switch msgID {
		case 0x01:
			fmt.Println("  Message: NAV-POSECEF (Position Solution in ECEF)")
		case 0x02:
			fmt.Println("  Message: NAV-POSLLH (Geodetic Position Solution)")
		case 0x03:
			fmt.Println("  Message: NAV-STATUS (Receiver Navigation Status)")
		case 0x04:
			fmt.Println("  Message: NAV-DOP (Dilution of Precision)")
		case 0x06:
			fmt.Println("  Message: NAV-SOL (Navigation Solution Information)")
		case 0x07:
			fmt.Println("  Message: NAV-PVT (Navigation Position Velocity Time Solution)")
		case 0x11:
			fmt.Println("  Message: NAV-VELECEF (Velocity Solution in ECEF)")
		case 0x12:
			fmt.Println("  Message: NAV-VELNED (Velocity Solution in NED)")
		case 0x20:
			fmt.Println("  Message: NAV-TIMEGPS (GPS Time Solution)")
		case 0x21:
			fmt.Println("  Message: NAV-TIMEUTC (UTC Time Solution)")
		case 0x30:
			fmt.Println("  Message: NAV-SVINFO (Space Vehicle Information)")
		default:
			fmt.Printf("  Message: Unknown NAV message ID 0x%02X\n", msgID)
		}
	case 0x02:
		classInfo = "RXM (Receiver Manager Messages)"
	case 0x05:
		classInfo = "ACK (Acknowledgement Messages)"
	case 0x06:
		classInfo = "CFG (Configuration Messages)"
	case 0x0A:
		classInfo = "MON (Monitoring Messages)"
	case 0x0B:
		classInfo = "AID (AssistNow Aiding Messages)"
	case 0x0D:
		classInfo = "TIM (Timing Messages)"
	case 0x10:
		classInfo = "ESF (External Sensor Fusion Messages)"
	case 0x13:
		classInfo = "MGA (Multiple GNSS Assistance Messages)"
	case 0x27:
		classInfo = "LOG (Logging Messages)"
	case 0xF0:
		classInfo = "SEC (Security Messages)"
	case 0xF1:
		classInfo = "HNR (High Rate Navigation Results)"
	default:
		classInfo = "Unknown Message Class"
	}

	fmt.Printf("  Class: %s (0x%02X)\n", classInfo, msgClass)
}

// changeBaudRate changes the baud rate of the serial connection
func changeBaudRate(port serial.Port, newBaudRate int) {
	// We need to store the port name before closing
	// Since we can't directly access the port name from the interface,
	// we'll need to ask the user for the port name again
	fmt.Println("To change the baud rate, please enter the port name:")

	// List available ports
	ports, err := listPorts()
	if err != nil {
		log.Printf("Error listing serial ports: %v", err)
		fmt.Println("Failed to change baud rate. Please restart the application.")
		return
	}

	if len(ports) == 0 {
		fmt.Println("No serial ports found. Please check your connections.")
		return
	}

	// Select port to use
	portName := selectPort(ports)

	// Close the current port
	port.Close()

	fmt.Printf("Changing baud rate to %d...\n", newBaudRate)

	// Reopen with new baud rate
	mode := &serial.Mode{
		BaudRate: newBaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	newPort, err := serial.Open(portName, mode)
	if err != nil {
		log.Printf("Error reopening port with new baud rate: %v", err)
		fmt.Println("Failed to change baud rate. Please restart the application.")
		return
	}

	// Set read timeout on new port
	err = newPort.SetReadTimeout(readTimeout)
	if err != nil {
		log.Printf("Warning: Error setting read timeout: %v", err)
	}

	// Replace the old port with the new one
	port = newPort

	fmt.Printf("Baud rate changed to %d successfully.\n", newBaudRate)
}
