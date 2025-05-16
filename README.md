# Go NTRIP Client for TOPGNSS TOP708

[![Go Tests](https://github.com/bramburn/go_ntrip/actions/workflows/go-test.yml/badge.svg)](https://github.com/bramburn/go_ntrip/actions/workflows/go-test.yml)
[![Go CI](https://github.com/bramburn/go_ntrip/actions/workflows/go-ci.yml/badge.svg)](https://github.com/bramburn/go_ntrip/actions/workflows/go-ci.yml)

A Go application for NTRIP client functionality and RTK position processing.

## Features

- Automatically detects and lists available serial ports
- Provides detailed information about USB devices (VID/PID)
- Interactive command interface for sending commands to the GNSS receiver
- Supports multiple data formats:
  - NMEA-0183 sentences with detailed parsing and display
  - RTCM3.3 messages for RTK corrections
  - u-blox UBX protocol messages
- NTRIP client functionality for connecting to NTRIP servers
- Built-in RTK processing for GNSS positioning
  - Position averaging for improved accuracy
  - Direct RTCM data processing without requiring external hardware
- Comprehensive error handling and troubleshooting tips
- Modular design following SOLID principles

## Project Structure

```
go_ntrip/
├── build/              # Build output directory
├── cmd/                # Application entry points
│   ├── gnss/           # Main GNSS application
│   ├── ntrip-client/   # NTRIP client application
│   ├── ntrip-avg/      # NTRIP position averaging application
│   ├── ntrip-rtk/      # NTRIP RTK processing application
│   ├── ntrip-server/   # NTRIP server application
│   └── relay/          # NTRIP relay application
├── internal/           # Private application code
│   ├── device/         # GNSS device communication
│   ├── ntrip/          # NTRIP client functionality
│   ├── parser/         # NMEA/RTCM/UBX parsers
│   ├── port/           # Serial port handling
│   ├── position/       # Position data handling
│   ├── rtk/            # RTK processing functionality
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
- `ntrip-pos` - Connect to NTRIP server and get fixed position
- `ntrip-avg` - Connect to NTRIP server and average position samples
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

## NTRIP Client Usage

The application includes NTRIP client functionality for connecting to NTRIP servers and obtaining RTK corrections.

### Using the CLI Commands

#### Getting a Single Fixed Position

When the main application is running, use the `ntrip-pos` command to connect to an NTRIP server and get a fixed position:

1. Enter the NTRIP server details when prompted (address, port, username, password, mountpoint)
2. The application will connect to the server and process RTCM data directly
3. Once an RTK fixed position is calculated, it will be saved to a JSON file in the application directory

#### Averaging Position Samples

Use the `ntrip-avg` command to connect to an NTRIP server and average multiple position samples for improved accuracy:

1. Enter the NTRIP server details when prompted (address, port, username, password, mountpoint)
2. Specify the minimum fix quality (4=RTK Fixed, 5=Float RTK) and number of samples to collect
3. The application will process RTCM data, calculate RTK solutions, and average the specified number of samples
4. The averaged position with statistics will be saved to a JSON file

### Using the Standalone NTRIP Applications

#### NTRIP Client for Single Position

You can use the standalone NTRIP client application to get a single fixed position:

```
go run cmd/ntrip-client/main.go -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -output base_position.json
```

Or with the built executable:

```
build/ntrip-client.exe -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH
```

#### NTRIP Position Averager

For more accurate positioning, use the position averaging application:

```
go run cmd/ntrip-avg/main.go -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -min-fix 4 -samples 60 -output base_position_avg.json
```

Or with the built executable:

```
build/ntrip-avg.exe -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -min-fix 4 -samples 60
```

#### NTRIP RTK Processor

For direct RTK processing of RTCM data:

```
go run cmd/ntrip-rtk/main.go -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -min-fix 4 -samples 60 -mode static -output rtk_position.json
```

Or with the built executable:

```
build/ntrip-rtk.exe -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -min-fix 4 -samples 60 -mode static
```

Command-line options:
- `-min-fix` - Minimum fix quality (4=RTK Fixed, 5=Float RTK)
- `-samples` - Number of position samples to collect and average
- `-timeout` - Maximum time to wait for samples (default: 10 minutes)
- `-mode` - Positioning mode (`static` or `kinematic`, default: `kinematic`)
  - Use `static` mode for base station setup to improve position stability
  - Use `kinematic` mode for rovers or when the receiver is moving

## RTK Implementation

The application implements Real-Time Kinematic (RTK) positioning using RTCM data from NTRIP servers. The RTK processor:

1. Parses RTCM messages from the data stream
2. Processes specific message types:
   - Message 1004: GPS L1/L2 observations
   - Message 1019: GPS ephemeris
   - Messages 1005/1006: Station coordinates (used for static mode)
3. Computes RTK solutions based on the collected data
4. Supports two positioning modes:
   - **Static Mode**: Optimized for base station setup, uses station coordinates when available
   - **Kinematic Mode**: Suitable for rovers or moving receivers

### Static Mode for Base Stations

When setting up a base station, use the static mode for improved position stability:

```
go run cmd/ntrip-rtk/main.go -address 192.168.0.64 -port 2101 -user reach -pass emlidreach -mount REACH -min-fix 4 -samples 60 -mode static
```

In static mode, the RTK processor:
- Uses RTCM station coordinates (message types 1005/1006) when available
- Applies more aggressive filtering to reduce position jitter
- Averages multiple samples for improved accuracy

## Future Development

- Configuration file support
- Data logging capabilities
- Support for additional GNSS receivers
- Web interface for monitoring and configuration
- Advanced RTK algorithms for improved accuracy

## License

MIT
