package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var log = logrus.New()

const version = `0.4.0`

var Version = &cli.Command{
	Name:         "version",
	Usage:        "get version",
	Description:  "getting application version",
	BashComplete: cli.DefaultAppComplete,
	Action:       actionVersion,
}

func actionVersion(ctx *cli.Context) error {
	fmt.Println(ctx.App.Name, version)
	return nil
}
