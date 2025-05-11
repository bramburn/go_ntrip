package parser

// UBXMessage represents a parsed UBX message
type UBXMessage struct {
	Class   byte   // Message class
	ID      byte   // Message ID
	Length  uint16 // Payload length
	Payload []byte // Message payload
	Valid   bool   // Whether the message is valid
}

// UBXParser provides functionality to parse UBX messages
type UBXParser struct {
	buffer []byte // Buffer to store partial messages
}

// NewUBXParser creates a new UBX parser
func NewUBXParser() *UBXParser {
	return &UBXParser{
		buffer: make([]byte, 0),
	}
}

// Process processes a chunk of data and extracts UBX messages
func (p *UBXParser) Process(data []byte) []UBXMessage {
	// Add new data to buffer
	p.buffer = append(p.buffer, data...)

	var messages []UBXMessage

	// Process UBX messages
	for len(p.buffer) >= 6 {
		// Check for UBX signature (0xB5 0x62)
		if p.buffer[0] == 0xB5 && p.buffer[1] == 0x62 {
			// Get message class and ID
			msgClass := p.buffer[2]
			msgID := p.buffer[3]

			// Get payload length (little endian)
			payloadLength := uint16(p.buffer[4]) | (uint16(p.buffer[5]) << 8)
			totalLength := int(payloadLength) + 8 // Add header and checksum

			if len(p.buffer) >= totalLength {
				// We have a complete message
				message := UBXMessage{
					Class:   msgClass,
					ID:      msgID,
					Length:  payloadLength,
					Payload: make([]byte, payloadLength),
					Valid:   true,
				}

				// Copy payload
				copy(message.Payload, p.buffer[6:6+payloadLength])

				// Add to result
				messages = append(messages, message)

				// Remove processed message from buffer
				p.buffer = p.buffer[totalLength:]
			} else {
				break // Wait for more data
			}
		} else {
			// Not a UBX message, skip this byte
			p.buffer = p.buffer[1:]
		}
	}

	return messages
}

// Reset clears the internal buffer
func (p *UBXParser) Reset() {
	p.buffer = p.buffer[:0]
}

// GetClassDescription returns a description of the UBX message class
func (p *UBXParser) GetClassDescription(msgClass byte) string {
	switch msgClass {
	case 0x01:
		return "NAV (Navigation Results)"
	case 0x02:
		return "RXM (Receiver Manager Messages)"
	case 0x05:
		return "ACK (Acknowledgement Messages)"
	case 0x06:
		return "CFG (Configuration Messages)"
	case 0x0A:
		return "MON (Monitoring Messages)"
	case 0x0B:
		return "AID (AssistNow Aiding Messages)"
	case 0x0D:
		return "TIM (Timing Messages)"
	case 0x10:
		return "ESF (External Sensor Fusion Messages)"
	case 0x13:
		return "MGA (Multiple GNSS Assistance Messages)"
	case 0x27:
		return "LOG (Logging Messages)"
	case 0xF0:
		return "SEC (Security Messages)"
	case 0xF1:
		return "HNR (High Rate Navigation Results)"
	default:
		return "Unknown Message Class"
	}
}

// GetMessageDescription returns a description of the UBX message
func (p *UBXParser) GetMessageDescription(msgClass byte, msgID byte) string {
	if msgClass == 0x01 { // NAV class
		switch msgID {
		case 0x01:
			return "NAV-POSECEF (Position Solution in ECEF)"
		case 0x02:
			return "NAV-POSLLH (Geodetic Position Solution)"
		case 0x03:
			return "NAV-STATUS (Receiver Navigation Status)"
		case 0x04:
			return "NAV-DOP (Dilution of Precision)"
		case 0x06:
			return "NAV-SOL (Navigation Solution Information)"
		case 0x07:
			return "NAV-PVT (Navigation Position Velocity Time Solution)"
		case 0x11:
			return "NAV-VELECEF (Velocity Solution in ECEF)"
		case 0x12:
			return "NAV-VELNED (Velocity Solution in NED)"
		case 0x20:
			return "NAV-TIMEGPS (GPS Time Solution)"
		case 0x21:
			return "NAV-TIMEUTC (UTC Time Solution)"
		case 0x30:
			return "NAV-SVINFO (Space Vehicle Information)"
		default:
			return "Unknown NAV message"
		}
	} else if msgClass == 0x06 { // CFG class
		switch msgID {
		case 0x00:
			return "CFG-PRT (Port Configuration)"
		case 0x01:
			return "CFG-MSG (Message Configuration)"
		case 0x04:
			return "CFG-RST (Reset Receiver)"
		case 0x24:
			return "CFG-NAV5 (Navigation Engine Settings)"
		case 0x31:
			return "CFG-TP5 (Time Pulse Parameters)"
		case 0x8A:
			return "CFG-VALDEL (Delete Configuration Values)"
		case 0x8B:
			return "CFG-VALGET (Get Configuration Values)"
		case 0x8C:
			return "CFG-VALSET (Set Configuration Values)"
		default:
			return "Unknown CFG message"
		}
	}

	return "Unknown message"
}
