package service

import (
	"log/slog"

	"github.com/uptrace/bun"
)

// Package-level dependencies. Initialized by Init().
var (
	db        *bun.DB
	logger    *slog.Logger //nolint:unused
	jwtSecret string
	apiKey    string
)

func init() {
	logger = slog.Default()
}

// Init sets up the service package dependencies.
func Init(database *bun.DB, log *slog.Logger, secret string, key string) {
	db = database
	if log != nil {
		logger = log
	}
	jwtSecret = secret
	apiKey = key
}

// Close releases the database connection.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
