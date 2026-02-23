package database

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mattn/go-sqlite3"
)

var (
	// ErrUnsupportedDriver is returned when the database driver is not recognized.
	ErrUnsupportedDriver = errors.New("unsupported database driver")
)

// IsDeadlockError checks if the error is a deadlock or "database busy" error.
func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrBusy ||
			sqliteErr.Code == sqlite3.ErrLocked
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001" || pgErr.Code == "40P01"
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
	}

	return false
}

// IsDuplicateKeyError checks if the error is a duplicate key / unique constraint violation.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	return false
}

// IsNotFoundError checks if the error indicates a row was not found.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}
