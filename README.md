# Go NTRIP Client for TOPGNSS TOP708

[![Go Tests](https://github.com/bramburn/go_ntrip/actions/workflows/go-test.yml/badge.svg)](https://github.com/bramburn/go_ntrip/actions/workflows/go-test.yml)
[![Go CI](https://github.com/bramburn/go_ntrip/actions/workflows/go-ci.yml/badge.svg)](https://github.com/bramburn/go_ntrip/actions/workflows/go-ci.yml)

A Go application for communicating with TOPGNSS TOP708 GNSS receivers via serial USB connection.

## Features

- Automatically detects and lists available serial ports
- Provides detailed information about USB devices (VID/PID)
- Interactive command interface for sending commands to the GNSS receiver
- Supports multiple data formats:
  - NMEA-0183 sentences with detailed parsing and display
  - RTCM3.3 messages for RTK corrections
  - u-blox UBX protocol messages
- Comprehensive error handling and troubleshooting tips
- Modular design following SOLID principles

## Project Structure

```
go_ntrip/
├── build/              # Build output directory
├── cmd/                # Application entry points
│   └── gnss/           # Main GNSS application
├── internal/           # Private application code
│   ├── device/         # GNSS device communication
│   ├── parser/         # NMEA/RTCM/UBX parsers
│   ├── port/           # Serial port handling
│   └── ui/             # User interface code
├── pkg/                # Public packages
├── scripts/            # Build scripts
└── test/               # Test files
```

## Prerequisites

- Go 1.16 or higher
- Connected TOPGNSS TOP708 GNSS receiver via USB

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

## Building

### Windows

Run the build script:

```
cd scripts
build.bat
```

The executable will be created in the `build` directory.

## Usage

Run the executable:

```
build/gnss_receiver.exe
```

Or run directly with Go:

```
go run cmd/gnss/main.go
```

### Commands

Once the application is running:

- `monitor` - Continuously display raw data from the receiver
- `nmea` - Monitor and parse NMEA sentences (GGA, RMC, GSV, GSA, GLL)
- `rtcm` - Monitor RTCM3.3 messages
- `ubx` - Monitor UBX protocol messages
- `baudrate <rate>` - Change the baud rate (e.g., `baudrate 115200`)
- `help` - Show available commands
- `exit` - Quit the application

You can also type any command to send directly to the GNSS receiver.

## TOPGNSS TOP708 Specifications

- Default baud rate: 38400 bps
- Supported protocols:
  - NMEA-0183 (standard sentences like $GNGGA, $GNGLL, $GNRMC)
  - RTCM3.3 for RTK corrections
  - u-blox UBX binary protocol
- Multi-constellation support: GPS, GLONASS, Galileo, BeiDou

## Configuration

The application is pre-configured for the TOPGNSS TOP708 with:

- Default baud rate: 38400 bps
- Port selection: Always prompts for selection at startup
- Read timeout: 500ms

## Common GNSS Commands

For TOPGNSS TOP708 receivers, you might use commands such as:

- `$PUBX,40,GGA,0,1,0,0,0,0*5B` - Enable NMEA GGA sentences
- `$PUBX,40,RMC,0,1,0,0,0,0*46` - Enable NMEA RMC sentences
- `$PUBX,40,GSV,0,1,0,0,0,0*59` - Enable NMEA GSV sentences

## Troubleshooting

If you encounter connection issues:

1. Verify the GNSS receiver is properly connected
2. Check that no other application is using the port
3. Try a different USB port
4. Ensure the correct drivers are installed
5. Try restarting the GNSS receiver

## Development

### Running Tests

```
go test ./test/...
```

### Continuous Integration

This project uses GitHub Actions for continuous integration:

- **Go Tests**: Runs all tests to ensure functionality works correctly
- **Go CI**: Performs linting, building, and test coverage reporting
- **Go Format**: Ensures code follows Go formatting standards

All workflows run automatically on push to main and on pull requests.

### Adding New Features

The modular design makes it easy to add new features:

1. To add a new parser, implement a new parser in the `internal/parser` package
2. To add a new device, implement the `GNSSDevice` interface in the `internal/device` package
3. To add a new UI, implement a new UI in the `internal/ui` package

## Future Development

- NTRIP client functionality for RTK corrections
- Configuration file support
- Data logging capabilities
- Support for additional GNSS receivers

## License

MIT
