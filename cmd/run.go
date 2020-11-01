package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/fs"
	"github.com/Meat-Hook/migrate/migrater"
	"github.com/urfave/cli/v2"
)

var Run = &cli.Command{
	Name:         "run",
	Usage:        "run migrate",
	Description:  "Migrate the DB To the most recent version available",
	BashComplete: cli.DefaultAppComplete,
	Before:       beforeMigrateAction,
	After:        afterMigrateAction,
	Action:       migrateAction,
	Flags:        dbFlags,
}

func beforeMigrateAction(ctx *cli.Context) error {
	log.Info("starting migration...")
	return nil
}

func afterMigrateAction(ctx *cli.Context) error {
	log.Info("finished")
	return nil
}

const (
	pg = `postgres`
	ms = `mysql`
)

func migrateAction(ctx *cli.Context) error {
	cmd, err := parse(ctx.String(Operation.Name))
	if err != nil {
		return err
	}

	dbDriver := ctx.String(Driver.Name)
	if dbDriver != pg && dbDriver != ms {
		return ErrUnknownDriver
	}

	dsn := make([]string, 0, 5)
	if ctx.String(Host.Name) != "" {
		dsn = append(dsn, fmt.Sprintf("host=%s", ctx.String(Host.Name)))
	}
	if ctx.Int(Port.Name) != 0 {
		dsn = append(dsn, fmt.Sprintf("port=%d", ctx.Int(Port.Name)))
	}
	if ctx.String(User.Name) != "" {
		dsn = append(dsn, fmt.Sprintf("user=%s", ctx.String(User.Name)))
	}
	if ctx.String(Pass.Name) != "" {
		dsn = append(dsn, fmt.Sprintf("password=%s", ctx.String(Pass.Name)))
	}
	if ctx.String(Name.Name) != "" {
		dsn = append(dsn, fmt.Sprintf("dbname=%s", ctx.String(Name.Name)))
	}
	dsn = append(dsn, "sslmode=disable")

	db, err := sql.Open(dbDriver, strings.Join(dsn, " "))
	if err != nil {
		return fmt.Errorf("open connect to database: %w", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Warnf("close connect to database: %s", err)
		}
	}()

	err = db.PingContext(ctx.Context)
	if err != nil {
		return fmt.Errorf("ping to database: %w", err)
	}

	m := migrater.New(db, log)
	filesSystem := fs.New()

	c := core.New(filesSystem, m)

	return c.Migrate(ctx.Context, ctx.String(Dir.Name), core.Config{
		Cmd: cmd,
		To:  ctx.Uint(To.Name),
	})
}

var (
	ErrUnknownOperation = errors.New("unknown Operation")
	ErrUnknownDriver    = errors.New("unknown Driver")
)

func parse(op string) (cmd core.MigrateCmd, err error) {
	switch op {
	case core.Up.String():
		cmd = core.Up
	case core.UpTo.String():
		cmd = core.UpTo
	case core.UpOne.String():
		cmd = core.UpOne
	case core.Down.String():
		cmd = core.Down
	case core.DownTo.String():
		cmd = core.DownTo
	case core.Reset.String():
		cmd = core.Reset
	default:
		return 0, ErrUnknownOperation
	}

	return cmd, nil
}
