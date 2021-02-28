package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/filesystem"
	"github.com/Meat-Hook/migrate/repo"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var Run = &cli.Command{
	Name:         "run",
	Usage:        "run migrate",
	Description:  "Migrate the DB To the most recent version available",
	BashComplete: cli.DefaultAppComplete,
	Action:       migrateAction,
	Flags:        []cli.Flag{Driver, Operation, To, Dir, DSN},
}

const (
	pg = `postgres`
	ms = `mysql`
)

func migrateAction(ctx *cli.Context) error {
	logger := zerolog.Ctx(ctx.Context)
	logger.Info().Msg("starting migration...")
	defer logger.Info().Msg("finished")

	cmd, err := parse(ctx.String(Operation.Name))
	if err != nil {
		return err
	}

	dbDriver := ctx.String(Driver.Name)
	if dbDriver != pg && dbDriver != ms {
		return ErrUnknownDriver
	}

	db, err := sql.Open(dbDriver, ctx.String(DSN.Name))
	if err != nil {
		return fmt.Errorf("open connect to database: %w", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error().Err(err).Msg("close connect to database")
		}
	}()

	err = db.PingContext(ctx.Context)
	if err != nil {
		return fmt.Errorf("ping to database: %w", err)
	}

	filesSystem := filesystem.New()

	tx, err := db.BeginTx(ctx.Context, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	r := repo.New(tx)
	c := core.New(*logger, filesSystem, r)

	err = c.Migrate(ctx.Context, ctx.String(Dir.Name), core.Config{
		Cmd: cmd,
		To:  ctx.Uint(To.Name),
	})
	if err != nil {
		errClose := tx.Rollback()
		if errClose != nil {
			logger.Error().Err(errClose).Msg("rollback")
		}

		return fmt.Errorf("migrate error: %w", err)
	}

	return tx.Commit()
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
	case core.DownTo.String():
		cmd = core.DownTo
	case core.Reset.String():
		cmd = core.Reset
	default:
		return 0, ErrUnknownOperation
	}

	return cmd, nil
}
