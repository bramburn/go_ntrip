package ntrip

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an NTRIP client
type Client struct {
	URL        string
	Username   string
	Password   string
	Mountpoint string
	httpClient *http.Client
}

// NewClient creates a new NTRIP client
func NewClient(url, username, password, mountpoint string) *Client {
	return &Client{
		URL:        url,
		Username:   username,
		Password:   password,
		Mountpoint: mountpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Connect connects to the NTRIP server and returns a reader for the RTCM data
func (c *Client) Connect(ctx context.Context) (io.ReadCloser, error) {
	// Create full URL with mountpoint if not already included
	fullURL := c.URL
	if c.Mountpoint != "" && !strings.Contains(fullURL, c.Mountpoint) {
		if !strings.HasSuffix(fullURL, "/") {
			fullURL += "/"
		}
		fullURL += c.Mountpoint
	}

	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set NTRIP specific headers
	req.Header.Set("User-Agent", "NTRIP go_ntrip/client")
	req.Header.Set("Ntrip-Version", "Ntrip/2.0")

	// Set authentication if provided
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NTRIP caster: %v", err)
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// Sourcetable represents an NTRIP sourcetable
type Sourcetable struct {
	Mounts []MountPoint
}

// MountPoint represents a mountpoint in an NTRIP sourcetable
type MountPoint struct {
	Name           string
	Identifier     string
	Format         string
	FormatDetails  string
	Carrier        int
	NavSystem      string
	Network        string
	Country        string
	Latitude       float64
	Longitude      float64
	NMEA           bool
	Solution       bool
	Generator      string
	Compression    string
	Authentication string
	Fee            bool
	Bitrate        int
}

// GetSourcetable retrieves the sourcetable from the NTRIP server
func (c *Client) GetSourcetable(ctx context.Context) (*Sourcetable, error) {
	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set NTRIP specific headers
	req.Header.Set("User-Agent", "NTRIP go_ntrip/client")
	req.Header.Set("Ntrip-Version", "Ntrip/2.0")

	// Set authentication if provided
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NTRIP caster: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the sourcetable
	sourcetable, err := parseSourcetable(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing sourcetable: %v", err)
	}

	return sourcetable, nil
}

// parseSourcetable parses a sourcetable string
func parseSourcetable(data string) (*Sourcetable, error) {
	lines := strings.Split(data, "\r\n")
	sourcetable := &Sourcetable{
		Mounts: []MountPoint{},
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "STR;") {
			fields := strings.Split(line, ";")
			if len(fields) < 15 {
				continue
			}

			mount := MountPoint{
				Name:          fields[1],
				Identifier:    fields[2],
				Format:        fields[3],
				FormatDetails: fields[4],
			}

			sourcetable.Mounts = append(sourcetable.Mounts, mount)
		}
	}

	return sourcetable, nil
}
