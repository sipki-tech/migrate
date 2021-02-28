package cmd

import (
	"fmt"

	"github.com/Meat-Hook/migrate/core"
	"github.com/urfave/cli/v2"
)

var (
	Driver = &cli.StringFlag{
		Name:       "driver",
		Aliases:    []string{"d"},
		Usage:      "database driver one of (postgres,mysql)",
		EnvVars:    []string{"DRIVER"},
		Value:      "postgres",
		Required:   true,
		HasBeenSet: true,
	}

	To = &cli.UintFlag{
		Name:    "to",
		Aliases: []string{"t"},
		Usage:   "on what element to migrate",
	}

	Operation = &cli.StringFlag{
		Name:     "operation",
		Aliases:  []string{"o"},
		Usage:    fmt.Sprintf("migration command one of (%s,%s,%s,%s)", core.Up, core.UpTo, core.DownTo, core.Reset),
		Required: true,
	}

	Dir = &cli.StringFlag{
		Name:    "dir",
		Aliases: []string{"D"},
		Usage:   "migration file location",
	}

	DSN = &cli.StringFlag{
		Name:     "dsn",
		Aliases:  []string{"D"},
		Usage:    "data source name for connection to database",
		EnvVars:  []string{"DSN"},
		Required: true,
	}

	MigrateName = &cli.StringFlag{
		Name:     "name",
		Aliases:  []string{"N"},
		Usage:    "migration file name",
		Required: true,
	}
)
