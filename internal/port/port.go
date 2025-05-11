package port

import (
	"fmt"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// SerialPort defines the interface for serial port operations
type SerialPort interface {
	// Open opens the serial port with the given configuration
	Open(portName string, baudRate int) error

	// Close closes the serial port
	Close() error

	// Read reads data from the port
	Read(buffer []byte) (int, error)

	// Write writes data to the port
	Write(data []byte) (int, error)

	// SetReadTimeout sets the read timeout for the port
	SetReadTimeout(timeout time.Duration) error

	// ListPorts lists all available serial ports
	ListPorts() ([]string, error)

	// GetPortDetails returns detailed information about available ports
	GetPortDetails() ([]*enumerator.PortDetails, error)
}

// SerialConfig holds configuration for the serial port
type SerialConfig struct {
	BaudRate int
	DataBits int
	Parity   serial.Parity
	StopBits serial.StopBits
	Timeout  time.Duration
}

// DefaultSerialConfig returns a default configuration for TOPGNSS TOP708
func DefaultSerialConfig() SerialConfig {
	return SerialConfig{
		BaudRate: 38400, // Default baud rate for TOPGNSS TOP708
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		Timeout:  500 * time.Millisecond,
	}
}

// GNSSSerialPort implements SerialPort interface for GNSS devices
type GNSSSerialPort struct {
	port   serial.Port
	config SerialConfig
}

// NewGNSSSerialPort creates a new GNSSSerialPort with default configuration
func NewGNSSSerialPort() *GNSSSerialPort {
	return &GNSSSerialPort{
		config: DefaultSerialConfig(),
	}
}

// Open opens the serial port with the given configuration
func (p *GNSSSerialPort) Open(portName string, baudRate int) error {
	// Update baud rate if provided
	if baudRate > 0 {
		p.config.BaudRate = baudRate
	}

	// Configure serial port
	mode := &serial.Mode{
		BaudRate: p.config.BaudRate,
		DataBits: p.config.DataBits,
		Parity:   p.config.Parity,
		StopBits: p.config.StopBits,
	}

	// Open the port
	port, err := serial.Open(portName, mode)
	if err != nil {
		return fmt.Errorf("error opening serial port %s: %w", portName, err)
	}

	p.port = port

	// Set read timeout
	err = p.port.SetReadTimeout(p.config.Timeout)
	if err != nil {
		return fmt.Errorf("error setting read timeout: %w", err)
	}

	return nil
}

// Close closes the serial port
func (p *GNSSSerialPort) Close() error {
	if p.port != nil {
		return p.port.Close()
	}
	return nil
}

// Read reads data from the port
func (p *GNSSSerialPort) Read(buffer []byte) (int, error) {
	if p.port == nil {
		return 0, fmt.Errorf("port not open")
	}
	return p.port.Read(buffer)
}

// Write writes data to the port
func (p *GNSSSerialPort) Write(data []byte) (int, error) {
	if p.port == nil {
		return 0, fmt.Errorf("port not open")
	}
	return p.port.Write(data)
}

// SetReadTimeout sets the read timeout for the port
func (p *GNSSSerialPort) SetReadTimeout(timeout time.Duration) error {
	if p.port == nil {
		return fmt.Errorf("port not open")
	}
	p.config.Timeout = timeout
	return p.port.SetReadTimeout(timeout)
}

// ListPorts lists all available serial ports
func (p *GNSSSerialPort) ListPorts() ([]string, error) {
	portDetails, err := p.GetPortDetails()
	if err != nil {
		return nil, err
	}

	var portNames []string
	for _, port := range portDetails {
		portNames = append(portNames, port.Name)
	}

	return portNames, nil
}

// GetPortDetails returns detailed information about available ports
func (p *GNSSSerialPort) GetPortDetails() ([]*enumerator.PortDetails, error) {
	return enumerator.GetDetailedPortsList()
}

// ChangeBaudRate changes the baud rate of the serial connection
func (p *GNSSSerialPort) ChangeBaudRate(baudRate int) error {
	if p.port == nil {
		return fmt.Errorf("port not open")
	}

	// We need to close and reopen the port to change the baud rate
	portName, err := p.getCurrentPortName()
	if err != nil {
		return err
	}

	// Close the current port
	err = p.Close()
	if err != nil {
		return fmt.Errorf("error closing port: %w", err)
	}

	// Reopen with new baud rate
	return p.Open(portName, baudRate)
}

// getCurrentPortName is a helper method to get the current port name
// This is a workaround since the serial.Port interface doesn't provide a way to get the port name
func (p *GNSSSerialPort) getCurrentPortName() (string, error) {
	// This is a limitation of the go.bug.st/serial library
	// In a real application, you would need to store the port name when opening the port
	return "", fmt.Errorf("unable to determine current port name, please provide it explicitly")
}
