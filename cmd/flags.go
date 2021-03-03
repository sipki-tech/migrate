package cmd

import (
	"github.com/urfave/cli/v2"
)

var (
	Dir = &cli.StringFlag{
		Name:    "dir",
		Aliases: []string{"D"},
		Usage:   "migration file location",
	}

	MigrateName = &cli.StringFlag{
		Name:     "name",
		Aliases:  []string{"N"},
		Usage:    "migration file name",
		Required: true,
	}
)
