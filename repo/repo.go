package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Meat-Hook/migrate/core"
)

var _ core.Repo = &Repo{}

// Repo implements core.Repo.
type Repo struct {
	tx *sql.Tx
}

// New returns new instance Repo.
func New(tx *sql.Tx) *Repo {
	return &Repo{tx: tx}
}

// Up for implements core.Repo.
func (r *Repo) Up(ctx context.Context, migrate core.Migrate) error {
	_, err := r.tx.ExecContext(ctx, migrate.Query.Up)
	if err != nil {
		return fmt.Errorf("up migrate %s: %w", migrate.Query.Up, err)
	}

	_, err = r.tx.ExecContext(ctx, "INSERT INTO migration (version) VALUES ($1)", migrate.Version)
	if err != nil {
		return fmt.Errorf("insert new version: %w", err)
	}

	return nil
}

// Rollback for implements core.Repo.
func (r *Repo) Rollback(ctx context.Context, migrate core.Migrate) error {
	_, err := r.tx.ExecContext(ctx, migrate.Query.Down)
	if err != nil {
		return fmt.Errorf("rollback migrate %s: %w", migrate.Query.Down, err)
	}

	_, err = r.tx.ExecContext(ctx, "DELETE FROM migration WHERE version = $1", migrate.Version)
	if err != nil {
		return fmt.Errorf("delete version: %w", err)
	}

	return nil
}

// Version for implements core.Repo.
func (r *Repo) Version(ctx context.Context) (uint, error) {
	const initTable = `CREATE TABLE IF NOT EXISTS migration
(
    id      serial,
    version integer                 not null,
    time    timestamp default now() not null,

    unique (version),
    primary key (id)
)`

	_, err := r.tx.ExecContext(ctx, initTable)
	if err != nil {
		return 0, fmt.Errorf("init default table: %w", err)
	}

	const query = `SELECT version FROM migration ORDER BY version DESC LIMIT 1`
	version := uint(0)
	err = r.tx.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("get current version: %w", err)
	}

	return version, nil
}
