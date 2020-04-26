package zergrepo

import "database/sql"

// TxOption type of parameters to start a transaction.
type TxOption func(options *sql.TxOptions)

// ReadOnly sets the transaction to be read-only.
func ReadOnly() TxOption {
	return func(options *sql.TxOptions) {
		options.ReadOnly = true
	}
}

// IsolationLevel sets the insulation level for the transaction.
func IsolationLevel(level sql.IsolationLevel) TxOption {
	return func(options *sql.TxOptions) {
		options.Isolation = level
	}
}
