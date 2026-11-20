package cli

import (
	"fmt"
	"log/slog"
	"os"

	"{{ cookiecutter.import_path }}/internal/config"
	"{{ cookiecutter.import_path }}/internal/service"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// migrationDir returns the dialect-specific migration directory.
func migrationDir(driver string) string {
	switch driver {
	case "postgres":
		return "migrations/pg"
	case "sqlite":
		return "migrations/sqlite"
	default:
		return "migrations"
	}
}

// gooseDialect maps our driver names to goose dialect constants.
func gooseDialect(driver string) goose.Dialect {
	switch driver {
	case "postgres":
		return goose.DialectPostgres
	case "sqlite":
		return goose.DialectSQLite3
	default:
		return goose.Dialect(driver)
	}
}

// openProvider loads config and creates a goose migration provider.
// Returns the provider, a cleanup function, and the config.
func openProvider(logger *slog.Logger) (*goose.Provider, *config.Config, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}

	conn, err := service.NewDBConn(cfg.Database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open database: %w", err)
	}

	dir := migrationDir(cfg.Database.Driver)
	provider, err := goose.NewProvider(gooseDialect(cfg.Database.Driver), conn, os.DirFS(dir))
	if err != nil {
		conn.Close()
		return nil, nil, nil, fmt.Errorf("create provider: %w", err)
	}

	cleanup := func() {
		provider.Close()
		conn.Close()
	}
	return provider, cfg, cleanup, nil
}

func newMigrateCommand(logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, cfg, cleanup, err := openProvider(logger)
			if err != nil {
				return err
			}
			defer cleanup()

			if _, err := provider.Up(cmd.Context()); err != nil {
				return fmt.Errorf("migrate up: %w", err)
			}

			logger.Info("migrations applied", slog.String("dir", migrationDir(cfg.Database.Driver)))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, _, cleanup, err := openProvider(logger)
			if err != nil {
				return err
			}
			defer cleanup()

			results, err := provider.Status(cmd.Context())
			if err != nil {
				return fmt.Errorf("migration status: %w", err)
			}

			for _, result := range results {
				logger.Info("migration",
					slog.Int64("version", result.Source.Version),
					slog.String("path", result.Source.Path),
					slog.String("state", string(result.State)),
				)
			}
			return nil
		},
	})

	return cmd
}
