package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v3"
	"github.com/urfave/cli/v3"
)

var version = "dev"

const (
	flagConfigPath = "config"
	flagStatePath  = "state"
)

var flags = cmd.Flags{
	&cli.StringFlag{
		Name:    flagConfigPath,
		Aliases: []string{"c"},
		Usage:   "Path to relay YAML config",
		Value:   "config.yaml",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagConfigPath)),
	},
	&cli.StringFlag{
		Name:     flagStatePath,
		Usage:    "Path to state file storing the last successful timestamp",
		Required: true,
		Sources:  cli.EnvVars(strcase.ToSNAKE(flagStatePath)),
	},
}.Merge(cmd.LogFlags)

func main() {
	os.Exit(realMain())
}

func realMain() (code int) {
	ui := newTerm()

	defer func() {
		if v := recover(); v != nil {
			ui.Error(fmt.Sprintf("Panic: %v\n%s", v, string(debug.Stack())))
			code = 1
		}
	}()

	app := cli.Command{
		Name:    "relay",
		Usage:   "Relay Aura-Mon energy data to configured databases",
		Version: version,
		Suggest: true,
		Flags:   flags,
		Action:  runRunner,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx, os.Args); err != nil {
		ui.Error(err.Error())
		return 1
	}
	return 0
}
