---
name: adding-cli-commands
description: "Use when adding, modifying, or debugging CLI commands with cobra. Don't use for web API endpoints or background tasks."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+, cobra"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Create command files in `internal/cli/` (one file per top-level command). Use `RunE` to return errors. Register commands in `init()` in `internal/cli/root.go`.

## Workflow

1. Create a new file in `internal/cli/` (e.g., `seed.go`)
2. Define a `*cobra.Command` variable with `Use`, `Short`, `Long`, `RunE`
3. Add flags in an `init()` function on the command file
4. Register in `root.go`'s `init()`: `rootCmd.AddCommand(newCmd)`
5. For subcommands: create a parent command, add children with `parentCmd.AddCommand(childCmd)`

### Adding a new command

```go
// internal/cli/seed.go
package cli

import (
    "fmt"
    "log/slog"

    "{{ cookiecutter.import_path }}/internal/config"
    "{{ cookiecutter.import_path }}/internal/service"

    "github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
    Use:   "seed",
    Short: "Seed the database with initial data",
    Long: `Seed the database with initial data for development.
This creates sample records for testing and demonstration.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.Load()
        if err != nil {
            return fmt.Errorf("load config: %w", err)
        }

        conn, err := service.NewDBConn(cfg.Database)
        if err != nil {
            return fmt.Errorf("open database: %w", err)
        }
        defer conn.Close()

        db := service.NewBunDB(cfg.Database, conn)
        logger := slog.Default()
        service.Init(db, logger, cfg.Auth.JwtSecret, cfg.Auth.APIKey)
        defer service.Close()

        if err := service.SeedData(); err != nil {
            return fmt.Errorf("seed data: %w", err)
        }

        fmt.Println("Database seeded successfully")
        return nil
    },
}

func init() {
    seedCmd.Flags().IntP("count", "c", 10, "Number of records to create")
    seedCmd.Flags().BoolP("force", "f", false, "Force re-seed even if data exists")
}
```

Then register in `internal/cli/root.go`:

```go
func init() {
    rootCmd.AddCommand(migrateCmd)
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(seedCmd)  // Add new command here
}
```

### Adding subcommands

```go
var dbCmd = &cobra.Command{
    Use:   "db",
    Short: "Database operations",
}

var dbMigrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations",
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
        return nil
    },
}

var dbResetCmd = &cobra.Command{
    Use:   "reset",
    Short: "Reset the database",
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
        return nil
    },
}

func init() {
    dbCmd.AddCommand(dbMigrateCmd)
    dbCmd.AddCommand(dbResetCmd)
    rootCmd.AddCommand(dbCmd)
}
```

This produces: `{{ cookiecutter.project_slug }} db migrate`, `{{ cookiecutter.project_slug }} db reset`.

### Adding flags

```go
func init() {
    // Integer flag with shorthand
    seedCmd.Flags().IntP("count", "c", 10, "Number of records to create")

    // Boolean flag with shorthand
    seedCmd.Flags().BoolP("force", "f", false, "Force re-seed even if data exists")

    // String flag
    seedCmd.Flags().StringP("output", "o", "json", "Output format (json, text)")
}

// Read flags in RunE
count, _ := cmd.Flags().GetInt("count")
force, _ := cmd.Flags().GetBool("force")
```

### Error handling

Always use `RunE` (not `Run`) to return errors. Wrap with context:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    conn, err := service.NewDBConn(cfg.Database)
    if err != nil {
        return fmt.Errorf("open database: %w", err)
    }
    defer conn.Close()
    // ...
    return nil
}
```

cobra prints the error and exits with code 1 automatically.

### Help text conventions

- **`Short`** — one-line description (shown in command list)
- **`Long`** — multi-line description (shown with `--help`)
- **Flag descriptions** — start with a verb, end with default value if applicable

```go
Use:   "migrate",
Short: "Run database migrations using goose",
Long: `Apply all pending database migrations.

This command connects to the configured database and runs
any migrations that have not yet been applied. Use --dry-run
to preview migrations without applying them.`,
```

## Database access

**CLI commands: create a separate bun connection for database access.**

`goose.Provider` doesn't expose its `*sql.DB`. CLI commands that need database access (not just migrations) should create their own bun connection:

```go
// In your CLI command's RunE:
cfg, err := config.Load()
// ... handle error ...

conn, err := service.NewDBConn(cfg.Database)
// ... handle error ...
defer conn.Close()

db := service.NewBunDB(cfg.Database, conn)
logger := slog.Default()
service.Init(db, logger, cfg.Auth.JwtSecret, cfg.Auth.APIKey)
defer service.Close()

// Now you can use service.ListXxx(), service.GetXxx(), etc.
```

## Key conventions

- Commands live in `internal/cli/` — one file per top-level command
- Use `RunE` for all commands — returns errors for proper exit codes
- Load config in each command — don't assume global state
- Register in `init()` of `root.go` — keeps all command wiring in one place
- Flags use `init()` function on the command — not in the struct literal
- Help text is descriptive — explain what the command does and any caveats
