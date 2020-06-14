// Package core contains business logic application.
package core

import (
	"context"
	"sort"
	"strings"
)

type (
	// Core manages the business logic of the application.
	Core interface {
		Migrate(ctx context.Context, dir string, cfg Config) error
		NewMigrate(_ context.Context, dir, name string) error
	}
	// Migrater manage migrate logic.
	Migrater interface {
		Migrate(ctx context.Context, cfg Config, m []Migrate) error
	}
	// Fs manages information from disk.
	Fs interface {
		Walk(dir string) ([]Migrate, error)
		CreateMigrate(dir, name string, m Migrate) error
	}
	// Migrate contains migrate information.
	Migrate struct {
		Version uint
		Query   Query
	}
	// Query contains `up` and `down` query.
	Query struct {
		Up   string
		Down string
	}
	core struct {
		fs Fs
		m  Migrater
	}
)

// New create new instance application.
func New(fs Fs, m Migrater) Core {
	return &core{
		fs: fs,
		m:  m,
	}
}

// Config migration configuration.
type Config struct {
	Cmd MigrateCmd
	To  uint
}

// MigrateCmd migration command.
type MigrateCmd uint8

// Migration command.
const (
	Up     MigrateCmd = iota + 1 // up
	UpOne                        // up-one
	UpTo                         // up-to
	Down                         // down
	DownTo                       // down-to
	Reset                        // reset
)

// Migrate migrates according to the settings.
func (a *core) Migrate(ctx context.Context, dir string, cfg Config) error {
	files, err := a.fs.Walk(dir)
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Version > files[j].Version
	})

	return a.m.Migrate(ctx, cfg, files)
}

// NewMigrate create new migrate files and add example query.
func (a *core) NewMigrate(_ context.Context, dir, name string) error {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}

	files, err := a.fs.Walk(dir)
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})

	const (
		exampleUp   = `create table test();`
		exampleDown = `drop table test;`
	)

	currentMaxVersion := uint(0)
	if len(files) > 0 {
		currentMaxVersion = files[len(files)-1].Version
	}
	newMigrate := Migrate{
		Version: currentMaxVersion + 1,
		Query: Query{
			Up:   exampleUp,
			Down: exampleDown,
		},
	}

	return a.fs.CreateMigrate(dir, name, newMigrate)
}
