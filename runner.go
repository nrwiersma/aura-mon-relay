package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/hamba/logger/v2"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/nrwiersma/aura-mon-relay/energy"
)

const interval = 5 * time.Second

// Client represents a source of energy metrics.
type Client interface {
	Get(ctx context.Context, start time.Time, intvl int) ([]energy.Row, error)
}

// DB represents a destination for energy metrics.
type DB interface {
	Write(ctx context.Context, metrics []database.Metric) error
}

// Storage represents a storage for the last successful timestamp.
type Storage interface {
	Read() (time.Time, error)
	Write(ts time.Time) error
}

// Runner is responsible for periodically fetching metrics from the client and
// sending them to the databases, while keeping track of the last successful
// timestamp using the storage.
type Runner struct {
	client    Client
	dbs       []DB
	storage   Storage
	initialTS time.Time

	log *logger.Logger
}

// NewRunner returns a relay runner.
func NewRunner(client Client, dbs []DB, storage Storage, initialTS time.Time) *Runner {
	return &Runner{
		client:    client,
		dbs:       dbs,
		storage:   storage,
		initialTS: initialTS,
	}
}

// Run periodically sends metrics from the client to the databases.
func (r *Runner) Run(ctx context.Context) error {
	lastTS := r.initialTS
	if storedTS, err := r.storage.Read(); err == nil && !storedTS.IsZero() {
		lastTS = storedTS
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		start := lastTS.Add(interval)
		rows, err := r.client.Get(ctx, start, int(interval.Seconds()))
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			metrics := toMetrics(rows)

			if err = r.sendMetrics(ctx, metrics); err != nil {
				return fmt.Errorf("sending metrics: %w", err)
			}

			lastTS = rows[len(rows)-1].Timestamp
			if err = r.storage.Write(lastTS); err != nil {
				return fmt.Errorf("writing to storage: %w", err)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Runner) sendMetrics(ctx context.Context, metrics []database.Metric) error {
	for _, db := range r.dbs {
		if err := db.Write(ctx, metrics); err != nil {
			return fmt.Errorf("writing to db: %w", err)
		}
	}
	return nil
}

func toMetrics(rows []energy.Row) []database.Metric {
	metrics := make([]database.Metric, 0, len(rows))
	for _, row := range rows {
		for _, device := range row.Devices {
			metrics = append(metrics, database.Metric{
				Measurement: device.Name,
				Timestamp:   row.Timestamp.Unix(),
				Fields: map[string]float64{
					"hz":           row.Hz,
					"volts":        device.Volts,
					"amps":         device.Amps,
					"watts":        device.Watts,
					"watt_hours":   device.WattHours,
					"power_factor": device.PowerFactor,
				},
			})
		}
	}
	return metrics
}
