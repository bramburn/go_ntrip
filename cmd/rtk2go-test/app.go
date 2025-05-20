package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// RTKApp represents the RTK application
type RTKApp struct {
	ntripClient       NTRIPClient
	gnssDevice        GNSSDevice
	rtkProcessor      RTKProcessor
	nmeaParser        NMEAParser
	status            RTKSolution
	statusMutex       sync.Mutex
	stopChan          chan struct{}
	colorOutput       bool
	reconnect         bool
	reconnectInterval int
	nmeaBuffer        string // Buffer to accumulate NMEA data across multiple reads
}

// RTKAppOptions represents options for the RTK application
type RTKAppOptions struct {
	ColorOutput       bool
	Reconnect         bool
	ReconnectInterval int
}

// NewRTKApp creates a new RTK application
func NewRTKApp(options RTKAppOptions) *RTKApp {
	return &RTKApp{
		stopChan: make(chan struct{}),
		status: RTKSolution{
			Status: rtkStatusNone,
			Time:   time.Now(),
		},
		colorOutput:       options.ColorOutput,
		reconnect:         options.Reconnect,
		reconnectInterval: options.ReconnectInterval,
		nmeaParser:        NewNMEAParser(),
	}
}

// SetGNSSDevice sets the GNSS device
func (app *RTKApp) SetGNSSDevice(device GNSSDevice) {
	app.gnssDevice = device
}

// SetNTRIPClient sets the NTRIP client
func (app *RTKApp) SetNTRIPClient(client NTRIPClient) {
	app.ntripClient = client
}

// SetRTKProcessor sets the RTK processor
func (app *RTKApp) SetRTKProcessor(processor RTKProcessor) {
	app.rtkProcessor = processor
}

// Start starts the RTK application
func (app *RTKApp) Start(logger Logger) error {
	// Validate required components
	if app.gnssDevice == nil {
		return fmt.Errorf("GNSS device not set")
	}

	if app.ntripClient == nil {
		return fmt.Errorf("NTRIP client not set")
	}

	if app.rtkProcessor == nil {
		return fmt.Errorf("RTK processor not set")
	}

	// Start connection monitoring if reconnection is enabled
	if app.reconnect {
		go app.monitorConnection(logger)
	}

	// Start monitoring solutions
	go app.monitorSolutions(logger, false)

	return nil
}

// Stop stops the RTK application
func (app *RTKApp) Stop(logger Logger) {
	logger.Println("Shutting down...")

	// Safely stop the RTK processor
	if app.rtkProcessor != nil {
		app.rtkProcessor.Stop()
	}

	// We're not calling Disconnect() on the NTRIP client because it causes a panic
	// The OS will clean up the resources when the process exits
}

// GetStatus returns the current RTK status
func (app *RTKApp) GetStatus() RTKSolution {
	app.statusMutex.Lock()
	defer app.statusMutex.Unlock()
	return app.status
}

// monitorConnection monitors the NTRIP connection and attempts to reconnect if it fails
func (app *RTKApp) monitorConnection(logger Logger) {
	ticker := time.NewTicker(time.Duration(app.reconnectInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			// Check if the client is connected
			if !app.ntripClient.IsConnected() {
				logger.Printf("NTRIP connection lost. Attempting to reconnect...")

				// Try to reconnect
				err := app.ntripClient.Connect()
				if err != nil {
					logger.Printf("Failed to reconnect to NTRIP server: %v", err)
					logger.Printf("Will retry in %d seconds...", app.reconnectInterval)
				} else {
					logger.Println("Successfully reconnected to NTRIP server.")

					// Restart the RTK processor if needed
					if app.rtkProcessor != nil {
						app.rtkProcessor.Stop()
						err = app.rtkProcessor.Start()
						if err != nil {
							logger.Printf("Failed to restart RTK processing: %v", err)
						} else {
							logger.Println("RTK processor restarted successfully.")
						}
					}
				}
			}
		}
	}
}

// monitorSolutions monitors RTK solutions and updates the status
func (app *RTKApp) monitorSolutions(logger Logger, verbose bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track status changes for notification
	var lastStatus string

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			// Get the current solution from the RTK processor
			sol := app.rtkProcessor.GetSolution()
			now := time.Now()

			// Update the status
			app.statusMutex.Lock()

			// Update status based on solution status
			app.status = sol
			app.status.Time = now

			// Read NMEA data directly from our GNSS device
			buffer := make([]byte, 4096) // Larger buffer to capture more data
			n, err := app.gnssDevice.ReadRaw(buffer)

			foundValidPosition := false

			if err != nil {
				if verbose {
					logger.Printf("Error reading from GNSS device: %v", err)
				}
			} else if n > 0 {
				if verbose {
					logger.Printf("Read %d bytes from GNSS device", n)
				}

				// Process NMEA data
				foundValidPosition = app.processNMEAData(buffer[:n], verbose, logger)
			}

			if !foundValidPosition {
				// If we couldn't parse any GGA sentences, use the RTK solution
				// but only if we don't already have a valid position
				if app.status.Latitude == 0 && app.status.Longitude == 0 {
					app.status = sol
					app.status.Time = now

					if verbose {
						logger.Printf("Using RTK solution as fallback: Lat: %f, Lon: %f", app.status.Latitude, app.status.Longitude)
					}
				} else {
					// Keep the last known position
					if verbose {
						logger.Printf("Keeping last known position: Lat: %f, Lon: %f", app.status.Latitude, app.status.Longitude)
					}
				}
			}

			// Check for status change
			statusChanged := lastStatus != app.status.Status
			lastStatus = app.status.Status

			app.statusMutex.Unlock()

			// Display status
			app.displayStatus(logger, statusChanged, verbose)
		}
	}
}

// processNMEAData processes NMEA data and updates the status
func (app *RTKApp) processNMEAData(data []byte, verbose bool, logger Logger) bool {
	// Append new data to the NMEA buffer
	app.nmeaBuffer += string(data)

	// Limit the buffer size to prevent memory issues
	if len(app.nmeaBuffer) > 16384 { // 16KB limit
		app.nmeaBuffer = app.nmeaBuffer[len(app.nmeaBuffer)-16384:]
	}

	if verbose {
		logger.Printf("NMEA buffer size: %d bytes", len(app.nmeaBuffer))
	}

	// Split the buffer into lines and look for complete NMEA sentences
	lines := strings.Split(app.nmeaBuffer, "\r\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "$") && strings.Contains(line, "GGA") {
			if verbose {
				logger.Printf("Found GGA sentence: %s", line)
			}

			// Parse GGA sentence
			ggaData, err := app.nmeaParser.ParseGGA(line)
			if err != nil {
				if verbose {
					logger.Printf("Error parsing GGA sentence: %v", err)
				}
				continue
			}

			// Update RTK status based on fix quality
			app.status.Status = GetFixQualityName(ggaData.Quality)

			// Update position
			app.status.Latitude = ggaData.Latitude
			app.status.Longitude = ggaData.Longitude
			app.status.Altitude = ggaData.Altitude
			app.status.NSats = ggaData.NumSats
			app.status.HDOP = ggaData.HDOP
			app.status.Age = ggaData.DGPSAge

			if verbose {
				logger.Printf("Parsed position: Lat: %f, Lon: %f, Alt: %f, Sats: %d, HDOP: %f, Age: %f",
					app.status.Latitude, app.status.Longitude, app.status.Altitude,
					app.status.NSats, app.status.HDOP, app.status.Age)
			}

			return true
		}
	}

	return false
}

// displayStatus displays the current status
func (app *RTKApp) displayStatus(logger Logger, statusChanged bool, verbose bool) {
	// Get status color
	statusColor := ""
	if app.colorOutput {
		switch app.status.Status {
		case rtkStatusNone:
			statusColor = colorRed
		case rtkStatusSingle:
			statusColor = colorYellow
		case rtkStatusFloat:
			statusColor = colorCyan
		case rtkStatusFix:
			statusColor = colorGreen
		}
	}

	// Format status with color if enabled
	var statusDisplay string
	if app.colorOutput {
		statusDisplay = fmt.Sprintf("%s%s%s%s%s",
			colorBold, statusColor, app.status.Status, colorReset, colorReset)
	} else {
		statusDisplay = app.status.Status
	}

	// Print status change notification
	if statusChanged {
		logger.Printf("RTK Status changed to: %s", statusDisplay)
	}

	// Print status
	statusStr := fmt.Sprintf("Status: %s | Lat: %.6f, Lon: %.6f, Alt: %.2fm | Sats: %d | Age: %.1fs",
		statusDisplay,
		app.status.Latitude,
		app.status.Longitude,
		app.status.Altitude,
		app.status.NSats,
		app.status.Age)

	// Print additional details if verbose
	if verbose {
		stats := app.rtkProcessor.GetStats()
		statusStr += fmt.Sprintf(" | Solutions: %d, Fix Ratio: %.2f%%",
			stats.Solutions,
			stats.FixRatio*100.0)
	}

	logger.Println(statusStr)
}
