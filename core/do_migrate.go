package core

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
)

type migrateOption struct {
	fs fs.FS
}

// MigrateOption option for migration.
type MigrateOption func(*migrateOption)

// WithCustomFS set custom filesystem.
func WithCustomFS(fs fs.FS) MigrateOption {
	return func(option *migrateOption) {
		option.fs = fs
	}
}

// Migrate migrates according to the settings.
func (c *Core) Migrate(ctx context.Context, dir string, cfg Config, options ...MigrateOption) error {
	opt := migrateOption{
		fs: c.fs,
	}

	for i := range options {
		options[i](&opt)
	}

	// collect migrates
	var migrates []Migrate
	err := c.fs.Walk(opt.fs, dir, func(path string, info fs.FileInfo) (err error) {
		if info.IsDir() {
			return nil
		}

		m, err := c.walkCb(path, info)
		if err != nil {
			return fmt.Errorf("path [%s]: call walk callback: %w", path, err)
		}

		migrates = append(migrates, *m)

		return nil
	})
	if err != nil {
		return err
	}

	switch cfg.Cmd {
	case Up:
		return c.Up(ctx, migrates...)
	case UpTo:
		return c.UpTo(ctx, cfg.To, migrates...)
	case DownTo:
		return c.DownTo(ctx, cfg.To, migrates...)
	case Reset:
		return c.Reset(ctx, migrates...)
	default:
		panic(fmt.Sprintf("unknown cmd: %d", cfg.Cmd))
	}
}

// Up performs all the migrations received.
func (c *Core) Up(ctx context.Context, migrates ...Migrate) error {
	return c.up(ctx, migrates...)
}

// UpTo migration to a specific version.
func (c *Core) UpTo(ctx context.Context, versionTo uint, migrates ...Migrate) error {
	upTo := make([]Migrate, 0, len(migrates))
	for i := range migrates {
		if migrates[i].Version <= versionTo {
			upTo = append(upTo, migrates[i])
		}
	}

	return c.up(ctx, upTo...)
}

// Reset rolls back all the migrations we've received.
func (c *Core) Reset(ctx context.Context, migrates ...Migrate) error {
	return c.down(ctx, migrates...)
}

// DownTo rollback to a specific version.
func (c *Core) DownTo(ctx context.Context, versionTo uint, migrates ...Migrate) error {
	downTo := make([]Migrate, 0, len(migrates))
	for i := range migrates {
		if migrates[i].Version >= versionTo {
			downTo = append(downTo, migrates[i])
		}
	}

	return c.down(ctx, downTo...)
}

func (c *Core) up(ctx context.Context, migrates ...Migrate) error {
	currentVersion, err := c.repo.Version(ctx)
	if err != nil {
		return err
	}

	err = validateMigrates(migrates...)
	if err != nil {
		return fmt.Errorf("validate migrates: %w", err)
	}

	sort.Slice(migrates, func(i, j int) bool {
		return migrates[i].Version < migrates[j].Version
	})

	for i := range migrates {
		if currentVersion >= migrates[i].Version {
			continue
		}

		err := c.repo.Up(ctx, migrates[i])
		if err != nil {
			return fmt.Errorf("call repo migrate: %w", err)
		}
	}

	return nil
}

func (c *Core) down(ctx context.Context, migrates ...Migrate) error {
	currentVersion, err := c.repo.Version(ctx)
	if err != nil {
		return err
	}

	err = validateMigrates(migrates...)
	if err != nil {
		return fmt.Errorf("validate migrates: %w", err)
	}

	sort.Slice(migrates, func(i, j int) bool {
		return migrates[i].Version > migrates[j].Version
	})

	for i := range migrates {
		if currentVersion < migrates[i].Version {
			continue
		}

		err := c.repo.Rollback(ctx, migrates[i])
		if err != nil {
			return fmt.Errorf("call repo migrate: %w", err)
		}
	}

	return nil
}

func validateMigrates(m ...Migrate) error {
	versions := map[uint]bool{}
	for i := range m {
		// Validate migrates.
		switch {
		case m[i].Version == 0:
			return fmt.Errorf("%w: you version %d", ErrNotCorrectVersion, m[i].Version)
		case m[i].Query.Up == "":
			return fmt.Errorf("%w: you version %d", ErrNotSetUp, m[i].Version)
		case m[i].Query.Down == "":
			return fmt.Errorf("%w: you version %d", ErrNotSetDown, m[i].Version)
		case versions[m[i].Version]:
			return ErrDuplicateVersion
		}

		versions[m[i].Version] = true
	}

	return nil
}
