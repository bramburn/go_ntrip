# RTK2go Test Client

This command-line application connects to RTK2go as an NTRIP client and uses a connected TOPGNSS TOP708 GNSS receiver to calculate RTK corrections. It displays real-time position and RTK status information.

## Features

- Connect to RTK2go NTRIP caster
- Process RTK corrections from various mountpoints
- Display real-time position and RTK status (NONE, SINGLE, DGPS, FLOAT, FIX)
- Monitor satellite count and correction age
- Support for different GNSS receiver baud rates
- Direct integration with TOPGNSS TOP708 GNSS receivers
- NMEA sentence parsing for accurate position information

## Usage

```bash
rtk2go-test -user your.email@example.com [options]
```

### Command-line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-server` | rtk2go.com | NTRIP server address |
| `-port` | 2101 | NTRIP server port |
| `-user` | (required) | NTRIP username (your email address) |
| `-password` | password | NTRIP password (any value for RTK2go) |
| `-mountpoint` | OCF-RH55LS-Capel | NTRIP mountpoint (OCF-RH55LS-Capel, MEDW, ozzy1) |
| `-gnss` | COM3 | GNSS receiver port |
| `-baud` | 38400 | GNSS receiver baud rate |
| `-duration` | 0 | Duration to run in seconds (0 for indefinite) |
| `-verbose` | false | Enable verbose output |

## Examples

### Basic Usage

```bash
# Connect to RTK2go using OCF-RH55LS-Capel mountpoint
rtk2go-test -user your.email@example.com -mountpoint OCF-RH55LS-Capel
```

### Testing Different Mountpoints

```bash
# Test with MEDW mountpoint
rtk2go-test -user your.email@example.com -mountpoint MEDW

# Test with ozzy1 mountpoint
rtk2go-test -user your.email@example.com -mountpoint ozzy1
```

### Specifying GNSS Receiver Settings

```bash
# Connect to GNSS receiver on COM4 with 115200 baud rate
rtk2go-test -user your.email@example.com -gnss COM4 -baud 115200
```

### Running for a Specific Duration

```bash
# Run for 5 minutes (300 seconds)
rtk2go-test -user your.email@example.com -duration 300
```

### Verbose Output

```bash
# Enable verbose output for additional statistics
rtk2go-test -user your.email@example.com -verbose
```

## RTK Status Explanation

The application displays the current RTK status:

- **NONE**: No position or invalid position
- **SINGLE**: Single solution (standard GNSS accuracy, ~2-5m)
- **DGPS**: Differential GPS solution (improved accuracy, ~1-3m)
- **FLOAT**: Float RTK solution (sub-meter accuracy, ~0.1-1m)
- **FIX**: Fixed RTK solution (centimeter accuracy, ~1-5cm)

The status is determined by parsing the GGA NMEA sentences from the GNSS receiver and examining the fix quality field.

## TOPGNSS TOP708 Device

This application is specifically designed to work with the TOPGNSS TOP708 GNSS receiver. The TOP708 is a high-precision multi-band GNSS receiver supporting RTK-based centimeter-level positioning.

### Device Features

- Multi-frequency GNSS board supporting GPS, BDS, GLONASS, GALILEO, and QZSS
- High-precision antenna
- RTK correction and RTCM support
- Mobile and base station operation
- Anti-interference and spoofing detection

### Connection Settings

- Default baud rate: 38400
- Data format: 8-N-1 (8 data bits, no parity, 1 stop bit)
- Typically connected via USB on COM3 or similar port

## Troubleshooting

1. **Connection Issues**:
   - Verify your internet connection
   - Check that the mountpoint is active on RTK2go
   - Ensure your email address is valid

2. **GNSS Receiver Issues**:
   - Verify the correct COM port
   - Check the baud rate settings (default is 38400)
   - Ensure the receiver is properly connected and powered
   - Make sure the device is outdoors with a clear view of the sky

3. **No Fix Status**:
   - Ensure your GNSS receiver has a clear view of the sky
   - Check that the selected mountpoint is within range (ideally <30km)
   - Allow sufficient time for convergence (can take several minutes)
   - Verify that the RTCM correction data is being received properly
