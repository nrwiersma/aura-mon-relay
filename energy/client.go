package energy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is an HTTP client for retrieving energy data from an Aura-Mon device.
type Client struct {
	client *http.Client
	url    *url.URL
}

// NewClient creates a new Client with the specified base URL.
func NewClient(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: u,
	}, nil
}

// Row is a single reading from the device, containing a timestamp, grid frequency,
// and per-device energy data.
type Row struct {
	Timestamp time.Time
	Hz        float64 // Grid frequency in hertz.
	Devices   []Device
}

// Device holds the energy measurements for a single monitored device.
type Device struct {
	Name        string
	Volts       float64 // Voltage in volts.
	Amps        float64 // Current in amperes.
	Watts       float64 // Real power in watts.
	WattHours   float64 // Energy consumed in watt-hours.
	PowerFactor float64 // Power factor (0–1).
}

// Get retrieves energy rows starting from start, covering one interval of intvl seconds.
func (c *Client) Get(ctx context.Context, start time.Time, intvl int) ([]Row, error) {
	v := url.Values{
		"start": []string{strconv.FormatInt(start.Unix(), 10)},
		"intvl": []string{strconv.Itoa(intvl)},
	}
	u := c.url.JoinPath("energy").ResolveReference(&url.URL{RawQuery: v.Encode()})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code for query %q: (%d) %s",
			req.URL.String(), resp.StatusCode, string(b),
		)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	rows, err := parseCSV(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("parsing CSV: %w", err)
	}
	return rows, nil
}
