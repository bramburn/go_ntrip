package ntrip

import (
	"fmt"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// RTKStats contains statistics about the RTK processing
type RTKStats struct {
	RoverObs  int     // Number of rover observations
	BaseObs   int     // Number of base observations
	Solutions int     // Number of solutions
	FixRatio  float64 // Ratio of fixed solutions
}

// RTKProcessor processes GNSS data using RTK
type RTKProcessor struct {
	receiver  *GNSSReceiver
	client    *Client
	svr       gnssgo.RtkSvr
	mutex     sync.Mutex
	running   bool
	solutions int
	fixCount  int
}

// NewRTKProcessor creates a new RTK processor
func NewRTKProcessor(receiver *GNSSReceiver, client *Client) (*RTKProcessor, error) {
	if receiver == nil {
		return nil, fmt.Errorf("receiver is nil")
	}
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	return &RTKProcessor{
		receiver: receiver,
		client:   client,
	}, nil
}

// Start starts the RTK processing
func (p *RTKProcessor) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return fmt.Errorf("already running")
	}

	// Configure RTK processing options
	var prcopt gnssgo.PrcOpt
	prcopt.Mode = gnssgo.PMODE_KINEMA               // Kinematic mode
	prcopt.NavSys = gnssgo.SYS_GPS | gnssgo.SYS_GLO // GPS + GLONASS
	prcopt.RefPos = 1                               // Use average of single position
	prcopt.ElMask = 15.0 * gnssgo.D2R               // Elevation mask (15 degrees)

	// Configure solution options
	var solopt [2]gnssgo.SolOpt
	solopt[0].Posf = gnssgo.SOLF_LLH  // Latitude/Longitude/Height format
	solopt[1].Posf = gnssgo.SOLF_NMEA // NMEA format

	// Configure stream types
	strtype := []int{
		gnssgo.STR_SERIAL,   // Rover input (physical GNSS receiver)
		gnssgo.STR_NTRIPCLI, // Base station input (NTRIP)
		gnssgo.STR_NONE,     // Ephemeris input
		gnssgo.STR_FILE,     // Solution 1 output
		gnssgo.STR_NONE,     // Solution 2 output
		gnssgo.STR_NONE,     // Log rover
		gnssgo.STR_NONE,     // Log base station
		gnssgo.STR_NONE,     // Log ephemeris
	}

	// Configure stream paths
	paths := []string{
		p.receiver.port, // Rover input (physical GNSS receiver)
		fmt.Sprintf("%s:%s@%s:%s/%s", p.client.username, p.client.password, p.client.server, p.client.port, p.client.mountpoint), // Base station input (NTRIP)
		"",                 // Ephemeris input
		"rtk_solution.pos", // Solution 1 output
		"",                 // Solution 2 output
		"",                 // Log rover
		"",                 // Log base station
		"",                 // Log ephemeris
	}

	// Configure stream formats
	strfmt := []int{
		gnssgo.STRFMT_UBX,   // Rover format (UBX)
		gnssgo.STRFMT_RTCM3, // Base station format (RTCM3)
		gnssgo.STRFMT_RINEX, // Ephemeris format
		gnssgo.SOLF_LLH,     // Solution 1 format
		gnssgo.SOLF_NMEA,    // Solution 2 format
	}

	// Start RTK server
	var errmsg string
	svrcycle := 10                        // Server cycle (ms)
	buffsize := 32768                     // Buffer size (bytes)
	navmsgsel := 0                        // Navigation message select
	cmds := []string{"", "", ""}          // Commands for input streams
	cmds_periodic := []string{"", "", ""} // Periodic commands
	rcvopts := []string{"", "", ""}       // Receiver options
	nmeacycle := 1000                     // NMEA request cycle (ms)
	nmeareq := 0                          // NMEA request type
	nmeapos := []float64{0, 0, 0}         // NMEA position

	// Start the RTK server
	if p.svr.RtkSvrStart(svrcycle, buffsize, strtype, paths, strfmt, navmsgsel,
		cmds, cmds_periodic, rcvopts, nmeacycle, nmeareq, nmeapos, &prcopt,
		solopt[:], nil, &errmsg) == 0 {
		return fmt.Errorf("failed to start RTK server: %s", errmsg)
	}

	p.running = true
	p.solutions = 0
	p.fixCount = 0

	// Start a goroutine to monitor solutions
	go p.monitorSolutions()

	return nil
}

// Stop stops the RTK processing
func (p *RTKProcessor) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	// Stop the RTK server
	cmds := []string{"", "", ""}
	p.svr.RtkSvrStop(cmds)

	p.running = false
	return nil
}

// GetStats returns statistics about the RTK processing
func (p *RTKProcessor) GetStats() RTKStats {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var sstat gnssgo.RtkSvrStat
	p.svr.RtkSvrGetStat(&sstat)

	fixRatio := 0.0
	if p.solutions > 0 {
		fixRatio = float64(p.fixCount) / float64(p.solutions)
	}

	return RTKStats{
		RoverObs:  sstat.Obs[0],
		BaseObs:   sstat.Obs[1],
		Solutions: p.solutions,
		FixRatio:  fixRatio,
	}
}

// monitorSolutions monitors the solutions produced by the RTK server
func (p *RTKProcessor) monitorSolutions() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mutex.Lock()
		if !p.running {
			p.mutex.Unlock()
			return
		}

		var sstat gnssgo.RtkSvrStat
		p.svr.RtkSvrGetStat(&sstat)

		// Check if we have a new solution
		if sstat.SolStat > 0 {
			p.solutions++
			if sstat.SolStat == gnssgo.SOLQ_FIX {
				p.fixCount++
			}
		}

		p.mutex.Unlock()
	}
}
