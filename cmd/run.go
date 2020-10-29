package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/fs"
	"github.com/Meat-Hook/migrate/migrater"
	"github.com/urfave/cli/v2"
)

var Run = &cli.Command{
	Name:         "run",
	Usage:        "run migrate",
	Description:  "Run the DB To the most recent version available",
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

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		ctx.String(Host.Name),
		ctx.Int(Port.Name),
		ctx.String(User.Name),
		ctx.String(Pass.Name),
		ctx.String(Name.Name),
	)

	db, err := sql.Open(dbDriver, dsn)
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
