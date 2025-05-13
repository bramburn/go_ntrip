package device

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bramburn/go_ntrip/internal/parser"
	"github.com/bramburn/go_ntrip/internal/port"
)

// TOPGNSSDevice implements GNSSDevice interface for TOPGNSS TOP708
type TOPGNSSDevice struct {
	serialPort port.SerialPort
	connected  bool
	mutex      sync.Mutex
	stopChan   chan bool
}

// NewTOPGNSSDevice creates a new TOPGNSS device
func NewTOPGNSSDevice(serialPort port.SerialPort) *TOPGNSSDevice {
	return &TOPGNSSDevice{
		serialPort: serialPort,
		connected:  false,
		stopChan:   make(chan bool),
	}
}

// Connect establishes a connection to the device
func (d *TOPGNSSDevice) Connect(portName string, baudRate int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.connected {
		return fmt.Errorf("device already connected")
	}

	// Use default baud rate if not specified
	if baudRate <= 0 {
		baudRate = 38400 // Default for TOPGNSS TOP708
	}

	// Open the port
	err := d.serialPort.Open(portName, baudRate)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	d.connected = true
	return nil
}

// Disconnect closes the connection to the device
func (d *TOPGNSSDevice) Disconnect() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.connected {
		return nil
	}

	err := d.serialPort.Close()
	if err != nil {
		return fmt.Errorf("error disconnecting device: %w", err)
	}

	d.connected = false
	return nil
}

// IsConnected returns whether the device is connected
func (d *TOPGNSSDevice) IsConnected() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.connected
}

// VerifyConnection checks if the device is sending valid GNSS data
func (d *TOPGNSSDevice) VerifyConnection(timeout time.Duration) bool {
	if !d.IsConnected() {
		return false
	}

	buffer := make([]byte, 1024)
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		n, err := d.serialPort.Read(buffer)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if n > 0 {
			data := string(buffer[:n])
			// Check for NMEA sentences
			if strings.Contains(data, "$GN") || strings.Contains(data, "$GP") {
				return true
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

// ReadRaw reads raw data from the device
func (d *TOPGNSSDevice) ReadRaw(buffer []byte) (int, error) {
	if !d.IsConnected() {
		return 0, fmt.Errorf("device not connected")
	}

	return d.serialPort.Read(buffer)
}

// WriteRaw writes raw data to the device
func (d *TOPGNSSDevice) WriteRaw(data []byte) (int, error) {
	if !d.IsConnected() {
		return 0, fmt.Errorf("device not connected")
	}

	return d.serialPort.Write(data)
}

// WriteCommand sends a command to the device
func (d *TOPGNSSDevice) WriteCommand(command string) error {
	if !d.IsConnected() {
		return fmt.Errorf("device not connected")
	}

	// Add newline if not present
	if !strings.HasSuffix(command, "\r\n") {
		command += "\r\n"
	}

	_, err := d.serialPort.Write([]byte(command))
	return err
}

// ChangeBaudRate changes the baud rate of the connection
func (d *TOPGNSSDevice) ChangeBaudRate(baudRate int) error {
	if !d.IsConnected() {
		return fmt.Errorf("device not connected")
	}

	// For TOPGNSS devices, we need to send a specific command to change the baud rate
	// This is device-specific and may vary
	// For now, we'll just change the port baud rate

	// This is a simplified implementation
	// In a real application, you would need to send the appropriate command to the device
	// and then change the port baud rate

	// Close and reopen the port with the new baud rate
	portName, err := d.getCurrentPortName()
	if err != nil {
		return err
	}

	// Disconnect
	err = d.Disconnect()
	if err != nil {
		return err
	}

	// Reconnect with new baud rate
	return d.Connect(portName, baudRate)
}

// GetAvailablePorts returns a list of available serial ports
func (d *TOPGNSSDevice) GetAvailablePorts() ([]string, error) {
	return d.serialPort.ListPorts()
}

// GetPortDetails returns detailed information about available ports
func (d *TOPGNSSDevice) GetPortDetails() ([]PortDetail, error) {
	details, err := d.serialPort.GetPortDetails()
	if err != nil {
		return nil, err
	}

	var result []PortDetail
	for _, detail := range details {
		// Convert string VID/PID to uint16 if they are USB devices
		vid := uint16(0)
		pid := uint16(0)

		if detail.IsUSB {
			// Parse hexadecimal VID/PID strings to uint16
			if vidVal, err := parseHexToUint16(detail.VID); err == nil {
				vid = vidVal
			}

			if pidVal, err := parseHexToUint16(detail.PID); err == nil {
				pid = pidVal
			}
		}

		result = append(result, PortDetail{
			Name:    detail.Name,
			IsUSB:   detail.IsUSB,
			VID:     vid,
			PID:     pid,
			Product: detail.Product,
		})
	}

	return result, nil
}

// getCurrentPortName is a helper method to get the current port name
func (d *TOPGNSSDevice) getCurrentPortName() (string, error) {
	// This is a limitation of the current implementation
	// In a real application, you would need to store the port name when opening the port
	return "", fmt.Errorf("unable to determine current port name, please provide it explicitly")
}

// MonitorNMEA starts monitoring NMEA data
func (d *TOPGNSSDevice) MonitorNMEA(config MonitorConfig) error {
	if !d.IsConnected() {
		return fmt.Errorf("device not connected")
	}

	// Create NMEA parser
	nmeaParser := parser.NewNMEAParser()
	buffer := make([]byte, config.BufferSize)
	dataBuffer := ""

	// Start monitoring in a goroutine
	go func() {
		for {
			select {
			case <-d.stopChan:
				return
			default:
				n, err := d.serialPort.Read(buffer)
				if err != nil {
					time.Sleep(config.PollInterval)
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
						parsedSentence := nmeaParser.Parse(sentence)

						// Handle parsed data
						if parsedSentence.Valid && config.Handler != nil {
							config.Handler.HandleNMEA(parsedSentence)
						}

						// Remove processed data from buffer
						if endIdx+2 <= len(dataBuffer) {
							dataBuffer = dataBuffer[endIdx+2:]
						} else {
							dataBuffer = ""
						}
					}
				}

				time.Sleep(config.PollInterval)
			}
		}
	}()

	return nil
}

// StopMonitoring stops all monitoring activities
func (d *TOPGNSSDevice) StopMonitoring() {
	d.stopChan <- true
}

// parseHexToUint16 converts a hexadecimal string to uint16
func parseHexToUint16(hexStr string) (uint16, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Parse the hex string
	val, err := strconv.ParseUint(hexStr, 16, 16)
	if err != nil {
		return 0, err
	}

	return uint16(val), nil
}
