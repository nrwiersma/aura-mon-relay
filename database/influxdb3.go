package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

// InfluxDB3Client is the interface required to interact with an InfluxDB 3.x client.
type InfluxDB3Client interface {
	WritePoints(ctx context.Context, points []*influxdb3.Point, options ...influxdb3.WriteOption) error
	Close() error
}

// InfluxDB3 implements the Database interface for InfluxDB 3.x.
type InfluxDB3 struct {
	client InfluxDB3Client
}

// NewInfluxDB3 creates a new InfluxDB3 instance with the given connection parameters.
func NewInfluxDB3(rawURL, token, database string) (*InfluxDB3, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid influxdb3 URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid influxdb3 URL %q: missing scheme or host", rawURL)
	}

	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     rawURL,
		Token:    token,
		Database: database,
	})
	if err != nil {
		return nil, fmt.Errorf("creating influxdb3 client: %w", err)
	}

	return NewInfluxDB3WithClient(client), nil
}

// NewInfluxDB3WithClient creates a new InfluxDB3 instance using an existing InfluxDB client.
func NewInfluxDB3WithClient(client InfluxDB3Client) *InfluxDB3 {
	return &InfluxDB3{
		client: client,
	}
}

// Write sends the given metrics to InfluxDB 3.x.
func (db *InfluxDB3) Write(ctx context.Context, metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	points := make([]*influxdb3.Point, 0, len(metrics))
	for _, metric := range metrics {
		point := influxdb3.NewPointWithMeasurement(metric.Measurement)
		for key, value := range metric.Tags {
			point.SetTag(key, value)
		}
		for key, value := range metric.Fields {
			point.SetField(key, value)
		}
		point.SetTimestamp(time.Unix(metric.Timestamp, 0).UTC())
		points = append(points, point)
	}

	if err := db.client.WritePoints(ctx, points); err != nil {
		return fmt.Errorf("writing metrics to influxdb3: %w", err)
	}
	return nil
}

// Close releases any resources held by the InfluxDB3 instance.
func (db *InfluxDB3) Close() error {
	return db.client.Close()
}
