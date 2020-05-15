package zergrepo

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
)

// Repo The wrapper around *sql.DB.
// Provides a number of convenient methods for starting a transaction
// and starting functions and wrapping returnable errors.
type Repo struct {
	db     *sqlx.DB
	log    Logger
	metric *Metric
	mapper Mapperer

	mu sync.Mutex // For migration management.
}

// Logger for logging :).
type Logger interface {
	Warnf(format string, args ...interface{})
}

// New return new instance Repo.
func New(db *sqlx.DB, log Logger, m *Metric, mapper Mapperer) *Repo {
	return &Repo{
		db:     db,
		log:    log,
		metric: m,
		mapper: mapper,
	}
}

// Close database connection.
func (r *Repo) Close() {
	r.WarnIfFail(r.db.Close)
}

// Tx automatically starts a transaction according to the parameters.
// If the callback returns the error, it will be wrapped and enriched with
// information about where the transaction was called from.
// Automatically collects metrics for function calls.
func (r *Repo) Tx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	methodName := callerMethodName()

	return r.metric.collect(methodName, func() error {
		return r.tx(ctx, methodName, nil, fn)
	})()
}

func (r *Repo) tx(ctx context.Context, methodName string, opts *sql.TxOptions, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", methodName, err)
	}

	err = fn(tx)
	if err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			err = fmt.Errorf("roolback err: %w", errRollback)
		}

		return fmt.Errorf("%s: %w", methodName, r.mapper.Map(err))
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", methodName, err)
	}

	return nil
}

// TxByCfg automatically starts a transaction according to the parameters.
// If the callback returns the error, it will be wrapped and enriched with
// information about where the transaction was called from.
// Automatically collects metrics for function calls.
func (r *Repo) TxByCfg(ctx context.Context, opts *sql.TxOptions, fn func(*sqlx.Tx) error) error {
	methodName := callerMethodName()

	return r.metric.collect(methodName, func() error {
		return r.tx(ctx, methodName, opts, fn)
	})()
}

// Do a wrapper for database requests.
// If the callback returns the error, it will be wrapped and enriched with
// information about where the transaction was called from.
// Automatically collects metrics for function calls.
func (r *Repo) Do(fn func(*sqlx.DB) error) error {
	methodName := callerMethodName()

	return r.metric.collect(methodName, func() error {
		err := fn(r.db)
		if err != nil {
			return fmt.Errorf("%s: %w", methodName, r.mapper.Map(err))
		}
		return nil
	})()
}
