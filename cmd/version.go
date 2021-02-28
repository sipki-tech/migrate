package cmd

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const version = `0.8.0`

var Version = &cli.Command{
	Name:         "version",
	Usage:        "cli version",
	Description:  "getting application version",
	BashComplete: cli.DefaultAppComplete,
	Action:       actionVersion,
}

func actionVersion(ctx *cli.Context) error {
	logger := zerolog.Ctx(ctx.Context)
	logger.Print(ctx.App.Name, version)
	return nil
}
