package test

import (
	"testing"
	"time"

	"github.com/bramburn/go_ntrip/internal/device"
	"github.com/bramburn/go_ntrip/internal/parser"
	"go.bug.st/serial/enumerator"
)

// MockSerialPort implements port.SerialPort for testing
type MockSerialPort struct {
	connected bool
	data      []byte
	written   []byte
}

func NewMockSerialPort() *MockSerialPort {
	return &MockSerialPort{
		connected: false,
		data:      []byte{},
		written:   []byte{},
	}
}

func (p *MockSerialPort) Open(portName string, baudRate int) error {
	p.connected = true
	return nil
}

func (p *MockSerialPort) Close() error {
	p.connected = false
	return nil
}

func (p *MockSerialPort) Read(buffer []byte) (int, error) {
	if !p.connected {
		return 0, nil
	}

	if len(p.data) == 0 {
		return 0, nil
	}

	n := copy(buffer, p.data)
	p.data = p.data[n:]
	return n, nil
}

func (p *MockSerialPort) Write(data []byte) (int, error) {
	if !p.connected {
		return 0, nil
	}

	p.written = append(p.written, data...)
	return len(data), nil
}

func (p *MockSerialPort) SetReadTimeout(timeout time.Duration) error {
	return nil
}

func (p *MockSerialPort) ListPorts() ([]string, error) {
	return []string{"COM1", "COM2"}, nil
}

func (p *MockSerialPort) GetPortDetails() ([]*enumerator.PortDetails, error) {
	return []*enumerator.PortDetails{
		{Name: "COM1", IsUSB: true, VID: 0x1234, PID: 0x5678, Product: "Mock GNSS"},
		{Name: "COM2", IsUSB: false},
	}, nil
}

// MockPortDetail implements port.PortDetail for testing
type MockPortDetail struct {
	Name    string
	IsUSB   bool
	VID     uint16
	PID     uint16
	Product string
}

// MockDataHandler implements device.DataHandler for testing
type MockDataHandler struct {
	nmeaMessages []parser.NMEASentence
	rtcmMessages []parser.RTCMMessage
	ubxMessages  []parser.UBXMessage
}

func NewMockDataHandler() *MockDataHandler {
	return &MockDataHandler{
		nmeaMessages: []parser.NMEASentence{},
		rtcmMessages: []parser.RTCMMessage{},
		ubxMessages:  []parser.UBXMessage{},
	}
}

func (h *MockDataHandler) HandleNMEA(sentence parser.NMEASentence) {
	h.nmeaMessages = append(h.nmeaMessages, sentence)
}

func (h *MockDataHandler) HandleRTCM(message parser.RTCMMessage) {
	h.rtcmMessages = append(h.rtcmMessages, message)
}

func (h *MockDataHandler) HandleUBX(message parser.UBXMessage) {
	h.ubxMessages = append(h.ubxMessages, message)
}

func TestTOPGNSSDevice(t *testing.T) {
	// Create mock serial port
	mockPort := NewMockSerialPort()

	// Create device with mock port
	gnssDevice := device.NewTOPGNSSDevice(mockPort)

	// Test connection
	err := gnssDevice.Connect("COM1", 38400)
	if err != nil {
		t.Errorf("Expected successful connection, got error: %v", err)
	}

	if !gnssDevice.IsConnected() {
		t.Error("Expected device to be connected")
	}

	// Test disconnection
	err = gnssDevice.Disconnect()
	if err != nil {
		t.Errorf("Expected successful disconnection, got error: %v", err)
	}

	if gnssDevice.IsConnected() {
		t.Error("Expected device to be disconnected")
	}

	// Test writing command
	err = gnssDevice.Connect("COM1", 38400)
	if err != nil {
		t.Errorf("Expected successful connection, got error: %v", err)
	}

	err = gnssDevice.WriteCommand("TEST")
	if err != nil {
		t.Errorf("Expected successful command write, got error: %v", err)
	}

	// Check that command was written with CRLF
	expected := "TEST\r\n"
	if string(mockPort.written) != expected {
		t.Errorf("Expected written data %q, got %q", expected, string(mockPort.written))
	}

	// Test reading data
	mockPort.data = []byte("$GNGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n")

	buffer := make([]byte, 1024)
	n, err := gnssDevice.ReadRaw(buffer)
	if err != nil {
		t.Errorf("Expected successful read, got error: %v", err)
	}

	if n != len(mockPort.data) {
		t.Errorf("Expected to read %d bytes, got %d", len(mockPort.data), n)
	}

	// Test verification
	mockPort.data = []byte("$GNGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n")

	if !gnssDevice.VerifyConnection(1 * time.Second) {
		t.Error("Expected successful verification")
	}

	// Test listing ports
	ports, err := gnssDevice.GetAvailablePorts()
	if err != nil {
		t.Errorf("Expected successful port listing, got error: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(ports))
	}

	// Test disconnection again
	err = gnssDevice.Disconnect()
	if err != nil {
		t.Errorf("Expected successful disconnection, got error: %v", err)
	}
}
