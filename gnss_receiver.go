package go_ntrip

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// GNSSReceiver represents a GNSS receiver that can generate test data
type GNSSReceiver struct {
	running   bool
	mutex     sync.Mutex
	dataQueue chan []byte
	stopChan  chan struct{}
}

// NewGNSSReceiver creates a new GNSS receiver simulator
func NewGNSSReceiver() *GNSSReceiver {
	return &GNSSReceiver{
		dataQueue: make(chan []byte, 100),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the GNSS receiver simulator
func (r *GNSSReceiver) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.running {
		return fmt.Errorf("receiver already running")
	}

	r.running = true
	go r.generateData()

	return nil
}

// Stop stops the GNSS receiver simulator
func (r *GNSSReceiver) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.running {
		return
	}

	r.running = false
	close(r.stopChan)
}

// Read reads data from the GNSS receiver
func (r *GNSSReceiver) Read(p []byte) (int, error) {
	select {
	case data, ok := <-r.dataQueue:
		if !ok {
			return 0, io.EOF
		}
		n := copy(p, data)
		return n, nil
	case <-time.After(1 * time.Second):
		return 0, fmt.Errorf("timeout waiting for data")
	}
}

// Write implements io.Writer but does nothing
func (r *GNSSReceiver) Write(p []byte) (int, error) {
	return len(p), nil
}

// Close closes the GNSS receiver
func (r *GNSSReceiver) Close() error {
	r.Stop()
	return nil
}

// generateData generates simulated GNSS data
func (r *GNSSReceiver) generateData() {
	// Generate data at 1Hz
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopChan:
			close(r.dataQueue)
			return
		case t := <-ticker.C:
			// Generate a simulated UBX NAV-PVT message
			data := r.generateUBXNavPVT(t)
			r.dataQueue <- data
		}
	}
}

// generateUBXNavPVT generates a simulated UBX NAV-PVT message
func (r *GNSSReceiver) generateUBXNavPVT(t time.Time) []byte {
	// UBX message header: 0xB5 0x62
	// Class: 0x01 (NAV)
	// ID: 0x07 (PVT)
	// Length: 92 bytes
	msg := []byte{
		0xB5, 0x62, // UBX header
		0x01, 0x07, // Class, ID
		0x5C, 0x00, // Length (92 bytes)
	}

	// Payload (92 bytes)
	payload := make([]byte, 92)

	// iTOW - GPS time of week (ms)
	itow := uint32(t.Unix() % (7 * 24 * 60 * 60) * 1000)
	payload[0] = byte(itow)
	payload[1] = byte(itow >> 8)
	payload[2] = byte(itow >> 16)
	payload[3] = byte(itow >> 24)

	// year, month, day
	payload[4] = byte(t.Year())
	payload[5] = byte(t.Year() >> 8)
	payload[6] = byte(t.Month())
	payload[7] = byte(t.Day())

	// hour, minute, second
	payload[8] = byte(t.Hour())
	payload[9] = byte(t.Minute())
	payload[10] = byte(t.Second())

	// valid flags
	payload[11] = 0x03 // validDate | validTime

	// tAcc - time accuracy estimate (ns)
	payload[12] = 0x00
	payload[13] = 0x00
	payload[14] = 0x00
	payload[15] = 0x00

	// nano - fraction of second (-500,000,000..500,000,000)
	nano := int32(t.Nanosecond())
	payload[16] = byte(nano)
	payload[17] = byte(nano >> 8)
	payload[18] = byte(nano >> 16)
	payload[19] = byte(nano >> 24)

	// fixType - GNSS fix type
	payload[20] = 0x03 // 3D fix

	// flags
	payload[21] = 0x01 // gnssFixOK

	// flags2
	payload[22] = 0x00

	// numSV - number of satellites used
	payload[23] = 0x08 // 8 satellites

	// lon - longitude (deg * 1e-7)
	// Using London coordinates as an example
	lon := int32(0.1276 * 1e7) // London longitude
	payload[24] = byte(lon)
	payload[25] = byte(lon >> 8)
	payload[26] = byte(lon >> 16)
	payload[27] = byte(lon >> 24)

	// lat - latitude (deg * 1e-7)
	lat := int32(51.5074 * 1e7) // London latitude
	payload[28] = byte(lat)
	payload[29] = byte(lat >> 8)
	payload[30] = byte(lat >> 16)
	payload[31] = byte(lat >> 24)

	// height - height above ellipsoid (mm)
	height := int32(10000) // 10 meters
	payload[32] = byte(height)
	payload[33] = byte(height >> 8)
	payload[34] = byte(height >> 16)
	payload[35] = byte(height >> 24)

	// hMSL - height above mean sea level (mm)
	hmsl := int32(9000) // 9 meters
	payload[36] = byte(hmsl)
	payload[37] = byte(hmsl >> 8)
	payload[38] = byte(hmsl >> 16)
	payload[39] = byte(hmsl >> 24)

	// hAcc - horizontal accuracy estimate (mm)
	hacc := uint32(1000) // 1 meter
	payload[40] = byte(hacc)
	payload[41] = byte(hacc >> 8)
	payload[42] = byte(hacc >> 16)
	payload[43] = byte(hacc >> 24)

	// vAcc - vertical accuracy estimate (mm)
	vacc := uint32(1500) // 1.5 meters
	payload[44] = byte(vacc)
	payload[45] = byte(vacc >> 8)
	payload[46] = byte(vacc >> 16)
	payload[47] = byte(vacc >> 24)

	// Add the payload to the message
	msg = append(msg, payload...)

	// Calculate checksum
	ck_a, ck_b := calculateUBXChecksum(msg[2:], len(payload)+4)
	msg = append(msg, ck_a, ck_b)

	return msg
}

// calculateUBXChecksum calculates the UBX checksum
func calculateUBXChecksum(data []byte, length int) (byte, byte) {
	var ck_a, ck_b byte
	for i := 0; i < length; i++ {
		ck_a += data[i]
		ck_b += ck_a
	}
	return ck_a, ck_b
}

// CreateGNSSStream creates a stream that uses the simulated GNSS receiver
func CreateGNSSStream() (*gnssgo.Stream, *GNSSReceiver, error) {
	// Create a new GNSS receiver
	receiver := NewGNSSReceiver()
	
	// Start the receiver
	if err := receiver.Start(); err != nil {
		return nil, nil, err
	}
	
	// Create a new stream
	stream := &gnssgo.Stream{}
	stream.InitStream()
	
	// Set the stream type to custom
	stream.Type = gnssgo.STR_SERIAL
	stream.Mode = gnssgo.STR_MODE_R
	stream.State = 1 // Open
	
	// Set the stream to use our receiver
	stream.Port = receiver
	
	return stream, receiver, nil
}
