package service

import (
	"database/sql"
	"fmt"
	"time"

	"{{ cookiecutter.import_path }}/internal/config"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	_ "modernc.org/sqlite"
)

// NewDBConn creates a raw sql.DB connection for the given config.
// Use this for operations that need direct database access (e.g., migrations).
func NewDBConn(cfg config.DatabaseConfig) (*sql.DB, error) {
	switch cfg.Driver {
	case "postgres":
		return sql.Open("pg", cfg.DSN)
	case "sqlite":
		return sql.Open("sqlite", cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// NewBunDB wraps a raw sql.DB with bun and applies connection pooling.
func NewBunDB(cfg config.DatabaseConfig, conn *sql.DB) *bun.DB {
	var db *bun.DB
	switch cfg.Driver {
	case "postgres":
		db = bun.NewDB(conn, pgdialect.New())
	case "sqlite":
		db = bun.NewDB(conn, sqlitedialect.New())
	default:
		db = bun.NewDB(conn, sqlitedialect.New())
	}

	// Apply connection pooling settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	return db
}
