package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type influxDB2Writer interface {
	WritePoint(ctx context.Context, point ...*write.Point) error
}

// InfluxDB2 writes metrics to an InfluxDB 2.x instance.
type InfluxDB2 struct {
	client influxdb2.Client
	writer influxDB2Writer
}

// NewInfluxDB2 creates a new InfluxDB2 using the given connection parameters.
func NewInfluxDB2(rawURL, token, org, bucket string) (*InfluxDB2, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid influxdb2 URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid influxdb2 URL %q: missing scheme or host", rawURL)
	}
	client := influxdb2.NewClient(rawURL, token)

	return NewInfluxDB2WithClient(client, org, bucket), nil
}

// NewInfluxDB2WithClient creates a new InfluxDB2 using the provided client.
func NewInfluxDB2WithClient(client influxdb2.Client, org, bucket string) *InfluxDB2 {
	return &InfluxDB2{
		client: client,
		writer: client.WriteAPIBlocking(org, bucket),
	}
}

// Write sends the given metrics to InfluxDB 2.x.
func (db *InfluxDB2) Write(ctx context.Context, metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	points := make([]*write.Point, 0, len(metrics))
	for _, metric := range metrics {
		points = append(points, write.NewPoint(
			metric.Measurement,
			metric.Tags,
			toInfluxFields(metric.Fields),
			time.Unix(metric.Timestamp, 0),
		))
	}

	if err := db.writer.WritePoint(ctx, points...); err != nil {
		return fmt.Errorf("writing metrics to influxdb2: %w", err)
	}
	return nil
}

func toInfluxFields(fields map[string]float64) map[string]any {
	influxFields := make(map[string]any, len(fields))
	for key, value := range fields {
		influxFields[key] = value
	}
	return influxFields
}

// Close releases any resources held by the InfluxDB2 instance.
func (db *InfluxDB2) Close() error {
	db.client.Close()
	return nil
}
