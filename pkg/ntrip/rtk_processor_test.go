package ntrip

import (
	"testing"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRtkSvr is a mock implementation of the gnssgo.RtkSvr type
type MockRtkSvr struct {
	mock.Mock
	RtkSvrStartCalled    bool
	RtkSvrStartArgs      []interface{}
	RtkSvrStartReturn    int
	RtkSvrStopCalled     bool
	RtkSvrStopArgs       []interface{}
	RtkSvrGetStatCalled  bool
	RtkSvrGetStatArgs    []interface{}
}

func (m *MockRtkSvr) RtkSvrStart(svrcycle, buffsize int, strtype []int, paths []string, strfmt []int, navsel int,
	cmds, cmds_periodic, rcvopts []string, nmeacycle, nmeareq int, nmeapos []float64, prcopt *gnssgo.PrcOpt,
	solopt []gnssgo.SolOpt, moni *gnssgo.Stream, errmsg *string) int {
	m.RtkSvrStartCalled = true
	m.RtkSvrStartArgs = []interface{}{svrcycle, buffsize, strtype, paths, strfmt, navsel,
		cmds, cmds_periodic, rcvopts, nmeacycle, nmeareq, nmeapos, prcopt, solopt, moni, errmsg}
	args := m.Called(svrcycle, buffsize, strtype, paths, strfmt, navsel,
		cmds, cmds_periodic, rcvopts, nmeacycle, nmeareq, nmeapos, prcopt, solopt, moni, errmsg)
	return args.Int(0)
}

func (m *MockRtkSvr) RtkSvrStop(cmds []string) {
	m.RtkSvrStopCalled = true
	m.RtkSvrStopArgs = []interface{}{cmds}
	m.Called(cmds)
}

func (m *MockRtkSvr) RtkSvrGetStat(stat *gnssgo.RtkSvrStat) {
	m.RtkSvrGetStatCalled = true
	m.RtkSvrGetStatArgs = []interface{}{stat}
	m.Called(stat)
	
	// Set some values in the stat object for testing
	stat.Obs[0] = 100 // Rover observations
	stat.Obs[1] = 50  // Base observations
	stat.SolStat = gnssgo.SOLQ_FIX // Solution status
}

// TestNewRTKProcessor tests the NewRTKProcessor function
func TestNewRTKProcessor(t *testing.T) {
	// Create mock receiver and client
	receiver := &GNSSReceiver{port: "COM1:9600:8:N:1", open: true}
	client := &Client{server: "example.com", port: "2101", username: "user", password: "pass", mountpoint: "MOUNT", connected: true}
	
	// Test with valid parameters
	processor, err := NewRTKProcessor(receiver, client)
	assert.NoError(t, err)
	assert.NotNil(t, processor)
	assert.Equal(t, receiver, processor.receiver)
	assert.Equal(t, client, processor.client)
	assert.False(t, processor.running)
	
	// Test with nil receiver
	processor, err = NewRTKProcessor(nil, client)
	assert.Error(t, err)
	assert.Nil(t, processor)
	assert.Contains(t, err.Error(), "receiver is nil")
	
	// Test with nil client
	processor, err = NewRTKProcessor(receiver, nil)
	assert.Error(t, err)
	assert.Nil(t, processor)
	assert.Contains(t, err.Error(), "client is nil")
}

// TestRTKProcessorStart tests the Start method
func TestRTKProcessorStart(t *testing.T) {
	// Create mock receiver and client
	receiver := &GNSSReceiver{port: "COM1:9600:8:N:1", open: true}
	client := &Client{
		server: "example.com", 
		port: "2101", 
		username: "user", 
		password: "pass", 
		mountpoint: "MOUNT", 
		connected: true,
	}
	
	// Create processor with mock RTK server
	processor := &RTKProcessor{
		receiver: receiver,
		client: client,
	}
	
	// Replace the RtkSvr with our mock
	mockRtkSvr := new(MockRtkSvr)
	processor.svr = mockRtkSvr
	
	// Setup mock expectations
	mockRtkSvr.On("RtkSvrStart", 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 
		mock.Anything).Return(1) // Return success
	
	// Test start
	err := processor.Start()
	assert.NoError(t, err)
	assert.True(t, processor.running)
	assert.True(t, mockRtkSvr.RtkSvrStartCalled)
	
	// Verify the paths were constructed correctly
	paths := mockRtkSvr.RtkSvrStartArgs[3].([]string)
	assert.Equal(t, "COM1:9600:8:N:1", paths[0]) // Rover input
	assert.Equal(t, "user:pass@example.com:2101/MOUNT", paths[1]) // Base station input
	
	// Test start when already running
	err = processor.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// TestRTKProcessorStop tests the Stop method
func TestRTKProcessorStop(t *testing.T) {
	// Create processor with mock RTK server
	processor := &RTKProcessor{
		running: true,
	}
	
	// Replace the RtkSvr with our mock
	mockRtkSvr := new(MockRtkSvr)
	processor.svr = mockRtkSvr
	
	// Setup mock expectations
	mockRtkSvr.On("RtkSvrStop", mock.Anything).Return()
	
	// Test stop
	err := processor.Stop()
	assert.NoError(t, err)
	assert.False(t, processor.running)
	mockRtkSvr.AssertCalled(t, "RtkSvrStop", []string{"", "", ""})
	
	// Test stop when not running
	err = processor.Stop()
	assert.NoError(t, err)
}

// TestRTKProcessorGetStats tests the GetStats method
func TestRTKProcessorGetStats(t *testing.T) {
	// Create processor with mock RTK server
	processor := &RTKProcessor{
		running: true,
		solutions: 10,
		fixCount: 5,
	}
	
	// Replace the RtkSvr with our mock
	mockRtkSvr := new(MockRtkSvr)
	processor.svr = mockRtkSvr
	
	// Setup mock expectations
	mockRtkSvr.On("RtkSvrGetStat", mock.Anything).Return()
	
	// Test GetStats
	stats := processor.GetStats()
	assert.Equal(t, 100, stats.RoverObs)
	assert.Equal(t, 50, stats.BaseObs)
	assert.Equal(t, 10, stats.Solutions)
	assert.Equal(t, 0.5, stats.FixRatio)
}
