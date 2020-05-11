package zergrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

// MigrateFunc migration function.
type MigrateFunc func(context.Context, *sql.Tx) error

// Migrate migration object, stores information about the migration version,
// as well as functions for their execution and rollback.
type Migrate struct {
	Version uint
	Up      MigrateFunc
	Down    MigrateFunc
}

// Query helper for convenient create MigrateFunc.
func Query(query string) MigrateFunc {
	return func(ctx context.Context, tx *sql.Tx) error {
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

var (
	ErrDuplicateVersion = errors.New("duplicate version")
)

// Up performs all the migrations received.
// Sorts in ascending order of versions if necessary.
func (r *Repo) Up(ctx context.Context, migrates ...Migrate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sort.Slice(migrates, func(i, j int) bool {
		return migrates[i].Version < migrates[j].Version
	})

	const getCurrentVersion = `SELECT version FROM migration ORDER BY version DESC LIMIT 1`

	return r.Tx(ctx, func(tx *sql.Tx) error {
		err := initTable.Up(ctx, tx)
		if err != nil {
			return fmt.Errorf("init table: %w", err)
		}

		currentVersion := uint(0)
		err = tx.QueryRowContext(ctx, getCurrentVersion).Scan(&currentVersion)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("get current version: %w", err)
		}

		versions := map[uint]bool{}
		for _, migrate := range migrates {
			switch {
			case currentVersion >= migrate.Version:
				continue
			case versions[migrate.Version]:
				return ErrDuplicateVersion
			}

			err = migrate.Up(ctx, tx)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "INSERT INTO migration (version) VALUES ($1)", migrate.Version)
			if err != nil {
				return fmt.Errorf("insert new version: %w", err)
			}

			versions[migrate.Version] = true
			currentVersion = migrate.Version
		}

		return nil
	})
}

// Down rolls back all the migrations we've received.
// Sorts in descending order of versions if necessary.
func (r *Repo) Down(ctx context.Context, migrates ...Migrate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sort.Slice(migrates, func(i, j int) bool {
		return migrates[j].Version < migrates[i].Version
	})

	const getCurrentVersion = `SELECT version FROM migration ORDER BY version DESC LIMIT 1`

	return r.Tx(ctx, func(tx *sql.Tx) error {
		currentVersion := uint(0)
		err := tx.QueryRowContext(ctx, getCurrentVersion).Scan(&currentVersion)
		switch {
		case err == sql.ErrNoRows:
			return nil
		case err != nil:
			return fmt.Errorf("get current version: %w", err)
		}

		versions := map[uint]bool{}
		for _, migrate := range migrates {
			switch {
			case currentVersion < migrate.Version:
				continue
			case versions[migrate.Version]:
				return ErrDuplicateVersion
			}

			err = migrate.Down(ctx, tx)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, "DELETE FROM migration WHERE version = $1", migrate.Version)
			if err != nil {
				return fmt.Errorf("insert new version: %w", err)
			}

			versions[migrate.Version] = true
			currentVersion = migrate.Version
		}

		return nil
	})
}
