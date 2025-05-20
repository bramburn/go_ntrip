package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bramburn/gnssgo/pkg/caster"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", 2101, "Port to listen on")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Configure logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create a new source service
	svc := caster.NewInMemorySourceService()

	// Add some example mounts to the sourcetable
	svc.Sourcetable = caster.Sourcetable{
		Casters: []caster.CasterEntry{
			{
				Host:       "localhost",
				Port:       *port,
				Identifier: "GNSSGO NTRIP Caster",
				Operator:   "GNSSGO",
				NMEA:       true,
				Country:    "USA",
				Latitude:   37.7749,
				Longitude:  -122.4194,
			},
		},
		Networks: []caster.NetworkEntry{
			{
				Identifier:          "GNSSGO",
				Operator:            "GNSSGO",
				Authentication:      "B",
				Fee:                 false,
				NetworkInfoURL:      "https://github.com/bramburn/gnssgo",
				StreamInfoURL:       "https://github.com/bramburn/gnssgo",
				RegistrationAddress: "admin@example.com",
			},
		},
		Mounts: []caster.StreamEntry{
			{
				Name:           "RTCM33",
				Identifier:     "RTCM33",
				Format:         "RTCM 3.3",
				FormatDetails:  "1004(1),1005/1006(5),1008(5),1012(1),1019(5),1020(5),1033(5),1042(5),1044(5),1045(5),1046(5)",
				Carrier:        "2",
				NavSystem:      "GPS+GLO+GAL+BDS+QZSS",
				Network:        "GNSSGO",
				CountryCode:    "USA",
				Latitude:       37.7749,
				Longitude:      -122.4194,
				NMEA:           true,
				Solution:       false,
				Generator:      "GNSSGO",
				Compression:    "none",
				Authentication: "B",
				Fee:            false,
				Bitrate:        9600,
			},
		},
	}

	// Create the caster
	caster := caster.NewCaster(fmt.Sprintf(":%d", *port), svc, logger)

	// Start the caster in a goroutine
	go func() {
		logger.Infof("Starting NTRIP caster on port %d", *port)
		if err := caster.ListenAndServe(); err != nil {
			logger.Fatalf("Caster error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown the caster
	logger.Info("Shutting down caster...")
	if err := caster.Shutdown(nil); err != nil {
		logger.Errorf("Error shutting down caster: %v", err)
	}
}
