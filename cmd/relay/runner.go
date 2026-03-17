package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/hamba/cmd/v3/observe"
	lctx "github.com/hamba/logger/v2/ctx"
	relay "github.com/nrwiersma/aura-mon-relay"
	"github.com/urfave/cli/v3"
)

func runRunner(ctx context.Context, cmd *cli.Command) error {
	obsvr, err := observe.New(ctx, cmd, "aura-mon-relay", &observe.Options{StatsRuntime: true})
	if err != nil {
		return err
	}
	defer obsvr.Close()

	obsvr.Log.Info("Aura Mon Relay Starting", lctx.Str("version", version))

	cfg, err := newConfig(cmd.String(flagConfigPath))
	if err != nil {
		return err
	}

	client, err := newClient(cfg.URL)
	if err != nil {
		return err
	}

	dbs, err := newDatabases(cfg.Databases)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := closeAll(dbs); closeErr != nil {
			obsvr.Log.Error("Error while closing database client(s)", lctx.Err(closeErr))
		}
	}()

	initialTS := cfg.InitialTS
	if initialTS.IsZero() {
		initialTS = time.Now().Add(-1 * time.Minute)
	}

	storage, err := newStorage(cmd.String(flagStatePath))
	if err != nil {
		return err
	}

	r := relay.NewRunner(client, dbs, storage, initialTS, obsvr)
	if err = r.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("runner error: %w", err)
	}
	return nil
}

func closeAll(dbs []relay.DB) error {
	var err error
	for _, db := range dbs {
		c, ok := db.(io.Closer)
		if !ok {
			continue
		}
		err = errors.Join(err, c.Close())
	}
	return err
}
