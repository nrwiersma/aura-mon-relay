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

// Row is a single data point for a device.
type Row struct {
	Timestamp time.Time
	Hz        float64
	Devices   []Device
}

// Device is a single device's energy data.
type Device struct {
	Name        string
	Volts       float64
	Amps        float64
	Watts       float64
	WattHours   float64
	PowerFactor float64
}

// Get retrieves energy data starting from the specified time with the given interval in seconds.
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
