package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hamba/cmd/v3/term"
	relay "github.com/nrwiersma/aura-mon-relay"
	"github.com/nrwiersma/aura-mon-relay/config"
	"github.com/nrwiersma/aura-mon-relay/database"
	"github.com/nrwiersma/aura-mon-relay/energy"
	"github.com/nrwiersma/aura-mon-relay/storage"
)

func newTerm() term.Term {
	return term.Prefixed{
		ErrorPrefix: "Error: ",
		Term: term.Colored{
			ErrorColor: term.Red,
			Term: term.Basic{
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
				Verbose:     false,
			},
		},
	}
}

func newConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Config{}, errors.New("config path required")
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return config.Config{}, fmt.Errorf("could not read config file %q: %w", path, err)
	}

	cfg, err := config.Parse(data)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not parse config file %q: %w", path, err)
	}

	if cfg.URL == "" {
		return config.Config{}, errors.New("config url required")
	}
	if len(cfg.Databases) == 0 {
		return config.Config{}, errors.New("at least one database config required")
	}

	return cfg, nil
}

func newClient(url string) (*energy.Client, error) {
	client, err := energy.NewClient(url)
	if err != nil {
		return nil, fmt.Errorf("could not create energy client: %w", err)
	}
	return client, nil
}

func newStorage(path string) (relay.Storage, error) {
	s, err := storage.NewFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not create file storage: %w", err)
	}
	return s, nil
}

func newDatabases(dbCfgs []config.Database) ([]relay.DB, error) {
	dbs := make([]relay.DB, 0, len(dbCfgs))
	for i, dbCfg := range dbCfgs {
		dbType := strings.ToLower(strings.TrimSpace(dbCfg.Type))

		switch dbType {
		case "influxdb2":
			if dbCfg.InfluxDB2 == nil {
				return nil, fmt.Errorf("database %d: influxdb2 config required", i)
			}

			db, err := database.NewInfluxDB2(
				dbCfg.InfluxDB2.URL,
				dbCfg.InfluxDB2.Token,
				dbCfg.InfluxDB2.Org,
				dbCfg.InfluxDB2.Bucket,
			)
			if err != nil {
				return nil, fmt.Errorf("database %d: creating influxdb2 client: %w", i, err)
			}

			dbs = append(dbs, db)
		case "influxdb3":
			if dbCfg.InfluxDB3 == nil {
				return nil, fmt.Errorf("database %d: influxdb3 config required", i)
			}

			db, err := database.NewInfluxDB3(
				dbCfg.InfluxDB3.URL,
				dbCfg.InfluxDB3.Token,
				dbCfg.InfluxDB3.Database,
			)
			if err != nil {
				return nil, fmt.Errorf("database %d: creating influxdb3 client: %w", i, err)
			}

			dbs = append(dbs, db)
		default:
			return nil, fmt.Errorf("database %d: unsupported database type %q", i, dbCfg.Type)
		}
	}
	return dbs, nil
}
