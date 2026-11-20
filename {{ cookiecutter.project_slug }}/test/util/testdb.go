package util

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	_ "modernc.org/sqlite"
)

// TestDB creates a database connection for testing.
// If APP_DATABASE_DRIVER=postgres and APP_DATABASE_DSN are set, connects to PostgreSQL.
// Otherwise, uses in-memory SQLite (default).
func TestDB(t *testing.T) *bun.DB {
	t.Helper()

	driver := os.Getenv("APP_DATABASE_DRIVER")
	dsn := os.Getenv("APP_DATABASE_DSN")

	if driver == "postgres" && dsn != "" {
		connector := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
		sqldb := sql.OpenDB(connector)

		db := bun.NewDB(sqldb, pgdialect.New())

		t.Cleanup(func() {
			sqldb.Close()
		})

		return db
	}

	// Default: in-memory SQLite
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create sqlite connection: %v", err)
	}

	db := bun.NewDB(conn, sqlitedialect.New())

	t.Cleanup(func() {
		conn.Close()
	})

	return db
}

// CreateTables creates tables for the given models in dependency order.
// Drops existing tables first to ensure a clean state (important for PostgreSQL).
// Pass model pointers in order: children before parents (reverse dependency order for drops).
//
//	usage:
//
//	    util.CreateTables(t, db, &model.Category{}, &model.Tag{}, &model.Page{}, &model.PageTag{})
func CreateTables(t *testing.T, db *bun.DB, models ...any) {
	t.Helper()

	// Drop existing tables first (reverse order to handle foreign keys)
	for i := len(models) - 1; i >= 0; i-- {
		m := models[i]
		_, err := db.NewDropTable().Model(m).IfExists().Exec(context.Background())
		if err != nil {
			t.Fatalf("failed to drop table for %T: %v", m, err)
		}
	}

	// Create tables in dependency order
	for _, m := range models {
		_, err := db.NewCreateTable().Model(m).Exec(context.Background())
		if err != nil {
			t.Fatalf("failed to create table for %T: %v", m, err)
		}
	}
}
