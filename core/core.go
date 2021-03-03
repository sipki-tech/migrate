package core

import (
	"context"
	"errors"
	"io/fs"

	"github.com/rs/zerolog"
)

//go:generate stringer --type MigrateCmd --linecomment --output stringer.go

// MigrateCmd migration command.
type MigrateCmd uint8

// Migration enum.
const (
	Up     MigrateCmd = iota + 1 // up
	UpTo                         // up-to
	DownTo                       // down-to
	Reset                        // reset
)

// Errors.
var (
	ErrDuplicateVersion  = errors.New("duplicate version")
	ErrNotCorrectVersion = errors.New("version must be above 0")
	ErrNotSetUp          = errors.New("not set up function")
	ErrNotSetDown        = errors.New("not set down function")
)

type (
	// FS manages file system.
	FS interface {
		fs.FS
		// Walk for walks on dir.
		Walk(string, func(string, fs.FileInfo) error) error
		// Mkdir make new dir if doesn't exist dir name.
		Mkdir(path string) error
		// SaveFile save new file.
		SaveFile(path string, buf []byte) error
	}

	// Repo provides to database.
	Repo interface {
		// Up makes migrate to database.
		Up(context.Context, Migrate) error
		// Rollback one migrate.
		Rollback(context.Context, Migrate) error
		// Version returns current migrate version from database.
		Version(context.Context) (uint, error)
	}

	// Migrate contains migrate information.
	Migrate struct {
		Version uint
		Query   Query
	}

	// Query contains `up` and `down` query.
	Query struct {
		Up   string `sql:"up"`
		Down string `sql:"down"`
	}

	// Config migration configuration.
	Config struct {
		Cmd MigrateCmd
		To  uint
	}

	// Core contains business logic for migrate database.
	Core struct {
		fs     FS
		logger zerolog.Logger
		repo   Repo
	}
)

// New builds and returns new instance business logic.
func New(logger zerolog.Logger, fs FS, repo Repo) *Core {
	return &Core{
		fs:     fs,
		logger: logger,
		repo:   repo,
	}
}
