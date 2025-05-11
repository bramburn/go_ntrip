package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bramburn/go_ntrip/internal/device"
	"github.com/bramburn/go_ntrip/internal/port"
	"github.com/bramburn/go_ntrip/internal/ui"
)

func main() {
	// Create serial port
	serialPort := port.NewGNSSSerialPort()

	// Create GNSS device
	gnssDevice := device.NewTOPGNSSDevice(serialPort)

	// Connect to device
	portName := selectPort(gnssDevice)
	if portName == "" {
		log.Fatal("No port selected. Exiting.")
	}

	fmt.Printf("Opening port %s with baud rate %d...\n", portName, 38400)
	err := gnssDevice.Connect(portName, 38400)
	if err != nil {
		handleConnectionError(err, portName)
		return
	}
	defer gnssDevice.Disconnect()

	fmt.Println("Port opened successfully. Waiting for device to initialize...")
	time.Sleep(2 * time.Second) // Give the device time to initialize

	// Verify connection
	if !gnssDevice.VerifyConnection(5 * time.Second) {
		fmt.Println("Unable to verify GNSS data. The device may not be sending data.")
		fmt.Println("Do you want to continue anyway? (y/n)")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Exiting...")
			return
		}
	}

	// Create and start CLI
	cli := ui.NewCLI(gnssDevice)
	cli.Start()
}

// selectPort prompts the user to select a port
func selectPort(device device.GNSSDevice) string {
	// List available ports
	ports, err := device.GetAvailablePorts()
	if err != nil {
		log.Fatalf("Error listing serial ports: %v", err)
	}

	if len(ports) == 0 {
		log.Fatal("No serial ports found. Please check your connections.")
	}

	// If only one port is available, use it
	if len(ports) == 1 {
		fmt.Printf("Only one port available. Using %s\n", ports[0])
		return ports[0]
	}

	// Get port details for better information
	details, err := device.GetPortDetails()
	if err != nil {
		// Fall back to simple list if details not available
		fmt.Println("Please select a port by number:")
		for i, port := range ports {
			fmt.Printf("%d: %s\n", i+1, port)
		}
	} else {
		fmt.Println("Available serial ports:")
		for i, detail := range details {
			portInfo := fmt.Sprintf("%d: %s", i+1, detail.Name)
			if detail.IsUSB {
				portInfo += fmt.Sprintf(" [USB: VID:%04X PID:%04X %s]",
					detail.VID, detail.PID, detail.Product)
			}
			fmt.Println(portInfo)
		}
	}

	// Prompt for selection
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter port number (or 0 to exit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var selection int
		_, err := fmt.Sscanf(input, "%d", &selection)
		if err == nil {
			if selection == 0 {
				return ""
			}
			if selection > 0 && selection <= len(ports) {
				return ports[selection-1]
			}
		}
		fmt.Println("Invalid selection. Please try again.")
	}
}

// handleConnectionError provides detailed error handling for connection issues
func handleConnectionError(err error, portName string) {
	log.Printf("Error opening serial port %s: %v", portName, err)

	fmt.Println("\nTroubleshooting tips:")
	fmt.Println("1. Check if the GNSS receiver is properly connected")
	fmt.Println("2. Verify that no other application is using the port")
	fmt.Println("3. Try a different USB port")
	fmt.Println("4. Check if the correct drivers are installed")
	fmt.Println("5. Try restarting the GNSS receiver")

	// Check for specific error types
	errStr := err.Error()
	if strings.Contains(errStr, "access denied") || strings.Contains(errStr, "permission denied") {
		fmt.Println("\nPermission issue detected:")
		fmt.Println("- Try running the application with administrator privileges")
		fmt.Println("- Check if another application is using the port")
	} else if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such file") {
		fmt.Println("\nPort not found:")
		fmt.Println("- The selected port may no longer be available")
		fmt.Println("- Try reconnecting the device and restarting the application")
	} else if strings.Contains(errStr, "timeout") {
		fmt.Println("\nConnection timeout:")
		fmt.Println("- The device is not responding")
		fmt.Println("- Check if the baud rate (38400) matches your device configuration")
	}
}
