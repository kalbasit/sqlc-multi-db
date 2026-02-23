package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kalbasit/sqlc-multi-db/example/pkg/database/mysqldb"
	"github.com/kalbasit/sqlc-multi-db/example/pkg/database/postgresdb"
	"github.com/kalbasit/sqlc-multi-db/example/pkg/database/sqlitedb"

	_ "github.com/go-sql-driver/mysql"      // MySQL driver
	_ "github.com/jackc/pgx/v5/stdlib"      // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"          // SQLite driver
)

// Open opens a database connection and returns a Querier.
// URL schemes: sqlite:/, postgresql://, mysql://
func Open(ctx context.Context, dbURL string) (Querier, error) {
	switch {
	case strings.HasPrefix(dbURL, "sqlite:"):
		path := strings.TrimPrefix(dbURL, "sqlite:")
		sdb, err := sql.Open("sqlite3", path)
		if err != nil {
			return nil, fmt.Errorf("opening sqlite: %w", err)
		}
		sdb.SetMaxOpenConns(1)
		if _, err := sdb.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
			return nil, fmt.Errorf("enabling foreign keys: %w", err)
		}
		return &sqliteWrapper{adapter: sqlitedb.NewAdapter(sdb)}, nil

	case strings.HasPrefix(dbURL, "postgres://"), strings.HasPrefix(dbURL, "postgresql://"):
		sdb, err := sql.Open("pgx", dbURL)
		if err != nil {
			return nil, fmt.Errorf("opening postgres: %w", err)
		}
		return &postgresWrapper{adapter: postgresdb.NewAdapter(sdb)}, nil

	case strings.HasPrefix(dbURL, "mysql://"):
		dsn := strings.TrimPrefix(dbURL, "mysql://")
		sdb, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("opening mysql: %w", err)
		}
		return &mysqlWrapper{adapter: mysqldb.NewAdapter(sdb)}, nil

	default:
		return nil, ErrUnsupportedDriver
	}
}
