// Package migrater contains logic for migrate data in database.
package migrater

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/Meat-Hook/migrate/core"
	"github.com/sirupsen/logrus"
)

var (
	ErrDuplicateVersion  = errors.New("duplicate version")
	ErrNotCorrectVersion = errors.New("version must be above 0")
	ErrNotSetUp          = errors.New("not set up function")
	ErrNotSetDown        = errors.New("not set down function")
)

// Migrater is responsible for data migration to the database.
type Migrater struct {
	l  logrus.FieldLogger
	db *sql.DB
}

// New create new instance migrater.
func New(db *sql.DB, l logrus.FieldLogger) *Migrater {
	return &Migrater{
		db: db,
		l:  l,
	}
}

// Migrate sql requests.
func (m *Migrater) Migrate(ctx context.Context, cfg core.Config, migrates []core.Migrate) error {
	switch cfg.Cmd {
	case core.Up:
		return m.Up(ctx, migrates...)
	case core.UpOne:
		return m.UpOne(ctx, migrates...)
	case core.UpTo:
		return m.UpTo(ctx, cfg.To, migrates...)
	case core.Down:
		return m.Down(ctx, migrates...)
	case core.DownTo:
		return m.DownTo(ctx, cfg.To, migrates...)
	case core.Reset:
		return m.Reset(ctx, migrates...)
	default:
		panic(fmt.Sprintf("unknown cmd: %d", cfg.Cmd))
	}
}

// Up performs all the migrations received.
func (m *Migrater) Up(ctx context.Context, migrates ...core.Migrate) error {
	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	return m.up(ctx, currentVersion, migrates...)
}

// UpTo migration to a specific version.
func (m *Migrater) UpTo(ctx context.Context, versionTo uint, migrates ...core.Migrate) error {
	upTo := make([]core.Migrate, 0, len(migrates))
	for i := range migrates {
		if migrates[i].Version <= versionTo {
			upTo = append(upTo, migrates[i])
		}
	}

	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	return m.up(ctx, currentVersion, upTo...)
}

// UpOne starting migration of the next version.
func (m *Migrater) UpOne(ctx context.Context, migrates ...core.Migrate) error {
	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	for i := range migrates {
		if currentVersion == 0 {
			return m.up(ctx, currentVersion, migrates[0])
		} else if migrates[i].Version == currentVersion {
			if len(migrates) > i+1 {
				return m.up(ctx, currentVersion, migrates[i+1])
			} else {
				return nil
			}
		}
	}

	return nil
}

func (m *Migrater) up(ctx context.Context, currentVersion uint, migrates ...core.Migrate) error {
	return m.tx(ctx, func(tx *sql.Tx) error {
		err := validateMigrates(migrates...)
		if err != nil {
			return err
		}

		sort.Slice(migrates, func(i, j int) bool {
			return migrates[i].Version < migrates[j].Version
		})

		for _, migrate := range migrates {
			if currentVersion >= migrate.Version {
				continue
			}

			err = Query(migrate.Query.Up)(ctx, tx)
			if err != nil {
				return fmt.Errorf("up %d: %w", migrate.Version, err)
			}
			m.l.Infof("up migrate: version %d", migrate.Version)
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
func (m *Migrater) Reset(ctx context.Context, migrates ...core.Migrate) error {
	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	return m.down(ctx, currentVersion, migrates...)
}

// Down rollback current migration.
func (m *Migrater) Down(ctx context.Context, migrates ...core.Migrate) error {
	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	for i := range migrates {
		if migrates[i].Version == currentVersion {
			return m.down(ctx, currentVersion, migrates[i])
		}
	}

	return nil
}

// DownTo rollback to a specific version.
func (m *Migrater) DownTo(ctx context.Context, versionTo uint, migrates ...core.Migrate) error {
	downTo := make([]core.Migrate, 0, len(migrates))
	for i := range migrates {
		if migrates[i].Version >= versionTo {
			downTo = append(downTo, migrates[i])
		}
	}

	currentVersion, err := m.currentVersion(ctx)
	if err != nil {
		return err
	}

	return m.down(ctx, currentVersion, downTo...)
}

func (m *Migrater) down(ctx context.Context, currentVersion uint, migrates ...core.Migrate) error {
	return m.tx(ctx, func(tx *sql.Tx) error {
		err := validateMigrates(migrates...)
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

			err = Query(migrate.Query.Down)(ctx, tx)
			if err != nil {
				return fmt.Errorf("down %d: %w", migrate.Version, err)
			}
			m.l.Infof("rollback migrate: version %d", migrate.Version)
			_, err = tx.ExecContext(ctx, "DELETE FROM migration WHERE version = $1", migrate.Version)
			if err != nil {
				return fmt.Errorf("delete version: %w", err)
			}

			currentVersion = migrate.Version
		}

		return nil
	})
}

func (m *Migrater) currentVersion(ctx context.Context) (version uint, err error) {
	const initTable = `CREATE TABLE IF NOT EXISTS migration
(
    id      serial,
    version integer                 not null,
    time    timestamp default now() not null,

    unique (version),
    primary key (id)
)`

	_, err = m.db.ExecContext(ctx, initTable)
	if err != nil {
		return 0, fmt.Errorf("init default table: %w", err)
	}

	const query = `SELECT version FROM migration ORDER BY time DESC LIMIT 1`
	err = m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("get current version: %w", err)
	}

	return version, nil
}

func (m *Migrater) tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			err = fmt.Errorf("%w and roolback err: %s", err, errRollback)
		}

		return fmt.Errorf("callback: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func validateMigrates(m ...core.Migrate) error {
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

// Query helper for convenient create MigrateFunc.
func Query(query string) func(ctx context.Context, tx *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("exec %s: %w", query, err)
		}
		return nil
	}
}
