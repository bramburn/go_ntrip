package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/hardware/topgnss/top708"
)

// TOP708Receiver implements the GNSSDevice interface
type TOP708Receiver struct {
	device    *top708.TOP708Device
	mutex     sync.Mutex
	connected bool
	portName  string
	baudRate  int
}

// NewTOP708Receiver creates a new TOP708Receiver
func NewTOP708Receiver(portName string, baudRate int) (*TOP708Receiver, error) {
	// Create a new serial port
	serialPort := top708.NewGNSSSerialPort()

	// Create a new TOP708 device
	device := top708.NewTOP708Device(serialPort)

	receiver := &TOP708Receiver{
		device:    device,
		connected: false,
		portName:  portName,
		baudRate:  baudRate,
	}

	return receiver, nil
}

// Connect connects to the GNSS receiver
func (r *TOP708Receiver) Connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connected {
		return fmt.Errorf("already connected")
	}

	// Connect to the device
	err := r.device.Connect(r.portName, r.baudRate)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	r.connected = true
	return nil
}

// Disconnect disconnects from the GNSS receiver
func (r *TOP708Receiver) Disconnect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return nil
	}

	// Disconnect from the device
	err := r.device.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from device: %w", err)
	}

	r.connected = false
	return nil
}

// IsConnected returns whether the receiver is connected
func (r *TOP708Receiver) IsConnected() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.connected && r.device.IsConnected()
}

// VerifyConnection checks if the device is sending valid GNSS data
func (r *TOP708Receiver) VerifyConnection(timeout time.Duration) bool {
	if !r.IsConnected() {
		return false
	}

	return r.device.VerifyConnection(timeout)
}

// Read implements the io.Reader interface
func (r *TOP708Receiver) Read(p []byte) (n int, err error) {
	if !r.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	return r.device.ReadRaw(p)
}

// Write implements the io.Writer interface
func (r *TOP708Receiver) Write(p []byte) (n int, err error) {
	if !r.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	return r.device.WriteRaw(p)
}

// ReadRaw reads raw data from the device
func (r *TOP708Receiver) ReadRaw(buffer []byte) (int, error) {
	if !r.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	return r.device.ReadRaw(buffer)
}

// WriteRaw writes raw data to the device
func (r *TOP708Receiver) WriteRaw(data []byte) (int, error) {
	if !r.IsConnected() {
		return 0, fmt.Errorf("not connected")
	}

	return r.device.WriteRaw(data)
}
