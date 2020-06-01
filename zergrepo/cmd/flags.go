package cmd

import (
	"fmt"

	zergrepo "github.com/ZergsLaw/zerg-repo"
	"github.com/ZergsLaw/zerg-repo/zergrepo/core"
	"github.com/urfave/cli/v2"
)

var (
	dbFlags = []cli.Flag{Driver, Operation, To, Dir, Name, User, Pass, Host, Port}
	Driver  = &cli.StringFlag{
		Name:       "db-driver",
		Aliases:    []string{"d"},
		Usage:      "database driver one of (postgres,mysql)",
		EnvVars:    []string{"DB_DRIVER"},
		Value:      pg,
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
		Usage:    fmt.Sprintf("migration command one of (%s,%s,%s,%s,%s,%s)", core.Up, core.UpTo, core.UpOne, core.Down, core.DownTo, core.Reset),
		Required: true,
	}
	Dir = &cli.StringFlag{
		Name:    "dir",
		Aliases: []string{"D"},
		Usage:   "migration file location",
	}
	Name = &cli.StringFlag{
		Name:       "db-name",
		Aliases:    []string{"n"},
		Usage:      "database name",
		EnvVars:    []string{"DB_NAME"},
		Value:      zergrepo.DBName,
		Required:   true,
		HasBeenSet: true,
	}
	User = &cli.StringFlag{
		Name:       "db-user",
		Aliases:    []string{"u"},
		Usage:      "database user",
		EnvVars:    []string{"DB_USER"},
		Value:      zergrepo.DBUser,
		Required:   true,
		HasBeenSet: true,
	}
	Pass = &cli.StringFlag{
		Name:       "db-pass",
		Aliases:    []string{"p"},
		Usage:      "database password",
		EnvVars:    []string{"DB_PASS"},
		Value:      zergrepo.DBPassword,
		Required:   true,
		HasBeenSet: true,
	}
	Host = &cli.StringFlag{
		Name:       "db-host",
		Aliases:    []string{"H"},
		Usage:      "database host",
		EnvVars:    []string{"DB_HOST"},
		Value:      zergrepo.DBHost,
		Required:   true,
		HasBeenSet: true,
	}
	Port = &cli.IntFlag{
		Name:       "db-port",
		Aliases:    []string{"P"},
		Usage:      "database port",
		EnvVars:    []string{"DB_PORT"},
		Value:      zergrepo.DBPort,
		Required:   true,
		HasBeenSet: true,
	}
	MigrateName = &cli.StringFlag{
		Name:     "migrate-name",
		Aliases:  []string{"M"},
		Usage:    "migration file name",
		Required: true,
	}
)
