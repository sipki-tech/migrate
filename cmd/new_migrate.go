package cmd

import (
	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/fs"
	"github.com/urfave/cli/v2"
)

var NewMigrate = &cli.Command{
	Name:         "create",
	Usage:        "create migrate file",
	Description:  "Creates a new migration file with test data.",
	BashComplete: cli.DefaultAppComplete,
	Action:       newMigrateAction,
	Flags:        []cli.Flag{MigrateName, Dir},
}

func newMigrateAction(ctx *cli.Context) error {
	filesSystem := fs.New()

	c := core.New(filesSystem, nil)

	return c.NewMigrate(ctx.Context, ctx.String(Dir.Name), ctx.String(MigrateName.Name))
}
