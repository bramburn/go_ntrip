# Go NTRIP Client

A Go application for communicating with GNSS receivers via serial USB connection.

## Features

- Automatically detects and lists available serial ports
- Provides detailed information about USB devices (VID/PID)
- Interactive command interface for sending commands to the GNSS receiver
- Continuous monitoring mode to observe GNSS data stream

## Prerequisites

- Go 1.16 or higher
- Connected GNSS receiver via USB

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/bramburn/go_ntrip.git
   cd go_ntrip
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Usage

Run the application:

```
go run main.go
```

### Commands

Once the application is running:

- Type commands to send directly to the GNSS receiver
- Type `monitor` to continuously display data from the receiver
- Type `exit` to quit the application

## Configuration

You can modify the following constants in `main.go` to match your GNSS receiver's requirements:

- `defaultBaudRate`: Set the baud rate (default: 9600)
- `defaultPort`: Specify a default port to use (leave empty to select at runtime)
- `readTimeout`: Adjust the read timeout (default: 500ms)

## Common GNSS Commands

Depending on your GNSS receiver, you might use commands such as:

- `$PMTK314,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0*28` - Enable NMEA GGA and RMC sentences
- `$PMTK220,1000*1F` - Set position update rate to 1Hz

Note: Commands may vary based on your specific GNSS receiver model.

## Future Development

- NTRIP client functionality for RTK corrections
- Configuration file support
- Data logging capabilities

## License

MIT
