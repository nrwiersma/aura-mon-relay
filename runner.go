// Package relay periodically fetches energy metrics from a client and forwards them
// to one or more databases, tracking progress with a persistent timestamp storage.
package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/hamba/cmd/v3/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/nrwiersma/aura-mon-relay/energy"
)

const interval = 5 * time.Second

// Client is the interface for fetching energy rows from a device.
type Client interface {
	Get(ctx context.Context, start time.Time, intvl int) ([]energy.Row, error)
}

// DB is the interface for writing metrics to a database.
type DB interface {
	Write(ctx context.Context, metrics []database.Metric) error
}

// Storage is the interface for persisting the last successfully processed timestamp.
type Storage interface {
	Read() (time.Time, error)
	Write(ts time.Time) error
}

// Runner periodically fetches metrics from a Client and writes them to one or more
// databases, tracking the last successful timestamp via Storage.
type Runner struct {
	client    Client
	dbs       []DB
	storage   Storage
	initialTS time.Time

	log *logger.Logger
}

// NewRunner returns a Runner that relays metrics from client to dbs.
func NewRunner(client Client, dbs []DB, storage Storage, initialTS time.Time, obsrv *observe.Observer) *Runner {
	return &Runner{
		client:    client,
		dbs:       dbs,
		storage:   storage,
		initialTS: initialTS,
		log:       obsrv.Log,
	}
}

// Run fetches and forwards metrics on a fixed interval until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	storedTS, err := r.storage.Read()
	if err != nil {
		return fmt.Errorf("reading last timestamp from storage: %w", err)
	}

	lastTS := r.initialTS
	if !storedTS.IsZero() {
		lastTS = storedTS
	}

	queryTS := lastTS.Add(interval)
	ts, n, err := r.collectAndSend(ctx, queryTS)
	if err != nil {
		return fmt.Errorf("relaying metrics: %w", err)
	}
	if !ts.IsZero() {
		lastTS = ts
	}

	r.log.Debug("Successfully relayed metrics",
		lctx.Time("timestamp", queryTS),
		lctx.Time("last", ts),
		lctx.Int("count", n),
	)

	nextCh := time.After(waitFor(n))
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-nextCh:
		}

		queryTS = lastTS.Add(interval)
		now := time.Now()
		limitTS := now.Truncate(interval).Add(-1 * interval)
		if queryTS.After(limitTS) || queryTS.Equal(limitTS) {
			nextCh = time.After(interval)
			continue
		}

		ts, n, err = r.collectAndSend(ctx, queryTS)
		if err != nil {
			r.log.Error("Could not relay metrics",
				lctx.Err(err),
				lctx.Time("timestamp", queryTS),
				lctx.Time("now", now),
			)

			nextCh = time.After(10 * time.Second)

			continue
		}
		nextCh = time.After(waitFor(n))

		if ts.IsZero() {
			if time.Until(lastTS).Abs() > time.Hour {
				lastTS = lastTS.Add(interval)
				r.log.Debug("No metrics found, but last timestamp is old. Advancing to next interval.",
					lctx.Time("last", lastTS),
				)
			}
			continue
		}
		lastTS = ts

		r.log.Debug("Successfully relayed metrics",
			lctx.Time("timestamp", queryTS),
			lctx.Time("last", ts),
			lctx.Int("count", n),
		)
	}
}

func (r *Runner) collectAndSend(ctx context.Context, start time.Time) (lastTS time.Time, n int, err error) {
	rows, err := r.client.Get(ctx, start, int(interval.Seconds()))
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("getting metrics: %w", err)
	}

	if len(rows) == 0 {
		return time.Time{}, 0, nil
	}

	metrics := toMetrics(rows)

	if err = r.sendMetrics(ctx, metrics); err != nil {
		return time.Time{}, 0, fmt.Errorf("sending metrics: %w", err)
	}

	lastTS = rows[len(rows)-1].Timestamp
	if err = r.storage.Write(lastTS); err != nil {
		return time.Time{}, 0, fmt.Errorf("writing to storage: %w", err)
	}

	return lastTS, len(rows), nil
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

func waitFor(n int) time.Duration {
	if n >= 100 || n <= 0 {
		return 100 * time.Millisecond
	}
	return interval
}
