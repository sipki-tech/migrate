package zergrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/jmoiron/sqlx"
)

// MigrateFunc migration function.
type MigrateFunc func(context.Context, *sqlx.Tx) error

// Migrate migration object, stores information about the migration version,
// as well as functions for their execution and rollback.
type Migrate struct {
	Version uint
	Up      MigrateFunc
	Down    MigrateFunc
}

var (
	registeredMigrates []Migrate
	mu                 sync.Mutex
)

var (
	ErrDuplicateVersion  = errors.New("duplicate version")
	ErrNotCorrectVersion = errors.New("version must be above 0")
	ErrNotSetUp          = errors.New("not set up function")
	ErrNotSetDown        = errors.New("not set down function")
)

// RegisterMetric records your migrations.
// Also validates for errors if multiple migrations have
// the same version or if no up or down logic was specified.
func RegisterMetric(m ...Migrate) error {
	mu.Lock()
	defer mu.Unlock()

	err := validateMigrates(m...)
	if err != nil {
		return err
	}

	registeredMigrates = append(registeredMigrates, m...)

	sort.Slice(registeredMigrates, func(i, j int) bool {
		return registeredMigrates[i].Version < registeredMigrates[j].Version
	})

	return nil
}

// Query helper for convenient create MigrateFunc.
func Query(query string) MigrateFunc {
	return func(ctx context.Context, tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("exec %s: %w", query, err)
		}
		return nil
	}
}

const (
	initVersion = 0

	upSQL = `CREATE TABLE IF NOT EXISTS migration
(
    id      serial,
    version integer                 not null,
    time    timestamp default now() not null,

    unique (version),
    primary key (id)
)`
	downSQL = `DROP TABLE migration;`
)

var (
	// migrate for create default table
	initTable = Migrate{
		Version: initVersion,
		Up:      Query(upSQL),
		Down:    Query(downSQL),
	}
)

// Up performs all the migrations received.
func (r *Repo) Up(ctx context.Context) error {
	return r.up(ctx, registeredMigrates...)
}

// UpTo migration to a specific version.
func (r *Repo) UpTo(ctx context.Context, versionTo uint) error {
	m := make([]Migrate, 0, len(registeredMigrates))
	for i := range registeredMigrates {
		if registeredMigrates[i].Version <= versionTo {
			m = append(m, registeredMigrates[i])
		}
	}

	return r.up(ctx, m...)
}

// UpOne starting migration of the next version.
func (r *Repo) UpOne(ctx context.Context) error {
	currentVersion, err := r.currentVersion(ctx)
	if err != nil {
		return err
	}

	for i := range registeredMigrates {
		if currentVersion == 0 {
			return r.up(ctx, registeredMigrates[0])
		} else if registeredMigrates[i].Version == currentVersion {
			if len(registeredMigrates) > i+1 {
				return r.up(ctx, registeredMigrates[i+1])
			} else {
				return nil
			}
		}
	}

	return nil
}

func (r *Repo) up(ctx context.Context, migrates ...Migrate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Tx(ctx, func(tx *sqlx.Tx) error {
		currentVersion, err := r.currentVersion(ctx)
		if err != nil {
			return err
		}

		err = validateMigrates(migrates...)
		if err != nil {
			return err
		}

		for _, migrate := range migrates {
			if currentVersion >= migrate.Version {
				continue
			}

			err = migrate.Up(ctx, tx)
			if err != nil {
				return fmt.Errorf("up %d: %w", migrate.Version, err)
			}
			r.log.Infof("up migrate: version %d", migrate.Version)
			_, err = tx.ExecContext(ctx, "INSERT INTO migration (version) VALUES ($1)", migrate.Version)
			if err != nil {
				return fmt.Errorf("insert new version: %w", err)
			}

			currentVersion = migrate.Version
		}

		return nil
	})
}

// Reset rolls back all the migrations we've received.
func (r *Repo) Reset(ctx context.Context) error {
	return r.down(ctx, registeredMigrates...)
}

// Down rollback current migration.
func (r *Repo) Down(ctx context.Context) error {
	currentVersion, err := r.currentVersion(ctx)
	if err != nil {
		return err
	}

	for i := range registeredMigrates {
		if registeredMigrates[i].Version == currentVersion {
			return r.down(ctx, registeredMigrates[i])
		}
	}

	return nil
}

// DownTo rollback to a specific version.
func (r *Repo) DownTo(ctx context.Context, versionTo uint) error {
	m := make([]Migrate, 0, len(registeredMigrates))
	for i := range registeredMigrates {
		if registeredMigrates[i].Version >= versionTo {
			m = append(m, registeredMigrates[i])
		}
	}

	return r.down(ctx, m...)
}

func (r *Repo) down(ctx context.Context, migrates ...Migrate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Tx(ctx, func(tx *sqlx.Tx) error {
		currentVersion, err := r.currentVersion(ctx)
		if err != nil {
			return err
		}

		err = validateMigrates(migrates...)
		if err != nil {
			return err
		}

		// reverse slice
		for i, j := 0, len(migrates)-1; i < j; i, j = i+1, j-1 {
			migrates[i], migrates[j] = migrates[j], migrates[i]
		}

		for _, migrate := range migrates {
			if currentVersion < migrate.Version {
				continue
			}

			err = migrate.Down(ctx, tx)
			if err != nil {
				return fmt.Errorf("down %d: %w", migrate.Version, err)
			}
			r.log.Infof("rollback migrate: version %d", migrate.Version)
			_, err = tx.ExecContext(ctx, "DELETE FROM migration WHERE version = $1", migrate.Version)
			if err != nil {
				return fmt.Errorf("delete version: %w", err)
			}

			currentVersion = migrate.Version
		}

		return nil
	})
}

func (r *Repo) currentVersion(ctx context.Context) (version uint, err error) {
	err = r.Tx(ctx, func(tx *sqlx.Tx) error {
		err := initTable.Up(ctx, tx)
		if err != nil {
			return fmt.Errorf("init table: %w", err)
		}

		const getCurrentVersion = `SELECT version FROM migration ORDER BY version DESC LIMIT 1`

		err = tx.QueryRowContext(ctx, getCurrentVersion).Scan(&version)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get current version: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return version, nil
}

func validateMigrates(m ...Migrate) error {
	versions := map[uint]bool{}
	for i := range m {
		// Validate migrates.
		switch {
		case m[i].Version == 0:
			return fmt.Errorf("%w: you version %d", ErrNotCorrectVersion, m[i].Version)
		case m[i].Up == nil:
			return fmt.Errorf("%w: you version %d", ErrNotSetUp, m[i].Version)
		case m[i].Down == nil:
			return fmt.Errorf("%w: you version %d", ErrNotSetDown, m[i].Version)
		case versions[m[i].Version]:
			return ErrDuplicateVersion
		}

		versions[m[i].Version] = true
	}

	return nil
}
