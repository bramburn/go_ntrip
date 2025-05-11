package parser

// RTCMMessage represents a parsed RTCM message
type RTCMMessage struct {
	MessageType int    // RTCM message type
	Length      int    // Message length
	Payload     []byte // Message payload
	Valid       bool   // Whether the message is valid
}

// RTCMParser provides functionality to parse RTCM messages
type RTCMParser struct {
	buffer []byte // Buffer to store partial messages
}

// NewRTCMParser creates a new RTCM parser
func NewRTCMParser() *RTCMParser {
	return &RTCMParser{
		buffer: make([]byte, 0),
	}
}

// Process processes a chunk of data and extracts RTCM messages
func (p *RTCMParser) Process(data []byte) []RTCMMessage {
	// Add new data to buffer
	p.buffer = append(p.buffer, data...)

	var messages []RTCMMessage

	// Process RTCM messages
	for len(p.buffer) >= 3 {
		// Check for RTCM signature (preamble byte 0xD3 followed by 2 bytes)
		if p.buffer[0] == 0xD3 {
			// Get message length from bytes 2-3 (10 bits)
			if len(p.buffer) < 6 {
				break // Not enough data yet
			}

			// Extract length (10 bits from bytes 1-2)
			length := (int(p.buffer[1]&0x03) << 8) | int(p.buffer[2])
			totalLength := length + 6 // Add header and CRC

			if len(p.buffer) >= totalLength {
				// We have a complete message
				messageType := (int(p.buffer[3]) << 4) | (int(p.buffer[4]) >> 4)

				// Create message
				message := RTCMMessage{
					MessageType: messageType,
					Length:      length,
					Payload:     make([]byte, length),
					Valid:       true,
				}

				// Copy payload
				copy(message.Payload, p.buffer[3:3+length])

				// Add to result
				messages = append(messages, message)

				// Remove processed message from buffer
				p.buffer = p.buffer[totalLength:]
			} else {
				break // Wait for more data
			}
		} else {
			// Not an RTCM message, skip this byte
			p.buffer = p.buffer[1:]
		}
	}

	return messages
}

// Reset clears the internal buffer
func (p *RTCMParser) Reset() {
	p.buffer = p.buffer[:0]
}

// GetMessageDescription returns a description of the RTCM message type
func (p *RTCMParser) GetMessageDescription(messageType int) string {
	switch messageType {
	case 1001:
		return "GPS L1-Only RTK Observables"
	case 1002:
		return "GPS Extended L1-Only RTK Observables"
	case 1003:
		return "GPS L1/L2 RTK Observables"
	case 1004:
		return "GPS Extended L1/L2 RTK Observables"
	case 1005:
		return "Stationary RTK Reference Station ARP"
	case 1006:
		return "Stationary RTK Reference Station ARP with Antenna Height"
	case 1007:
		return "Antenna Descriptor"
	case 1008:
		return "Antenna Descriptor & Serial Number"
	case 1009:
		return "GLONASS L1-Only RTK Observables"
	case 1010:
		return "GLONASS Extended L1-Only RTK Observables"
	case 1011:
		return "GLONASS L1/L2 RTK Observables"
	case 1012:
		return "GLONASS Extended L1/L2 RTK Observables"
	case 1019:
		return "GPS Ephemerides"
	case 1020:
		return "GLONASS Ephemerides"
	case 1033:
		return "Receiver and Antenna Descriptors"
	case 1071, 1072, 1073, 1074, 1075, 1076, 1077:
		return "GPS MSM" + string(rune(messageType-1070+'0'))
	case 1081, 1082, 1083, 1084, 1085, 1086, 1087:
		return "GLONASS MSM" + string(rune(messageType-1080+'0'))
	case 1091, 1092, 1093, 1094, 1095, 1096, 1097:
		return "Galileo MSM" + string(rune(messageType-1090+'0'))
	case 1101, 1102, 1103, 1104, 1105, 1106, 1107:
		return "SBAS MSM" + string(rune(messageType-1100+'0'))
	case 1111, 1112, 1113, 1114, 1115, 1116, 1117:
		return "QZSS MSM" + string(rune(messageType-1110+'0'))
	case 1121, 1122, 1123, 1124, 1125, 1126, 1127:
		return "BeiDou MSM" + string(rune(messageType-1120+'0'))
	default:
		return "Unknown RTCM Message Type"
	}
}
