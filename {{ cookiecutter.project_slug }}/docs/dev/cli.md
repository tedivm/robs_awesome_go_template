# CLI — Cobra

{{ cookiecutter.project_name }} uses [cobra](https://github.com/spf13/cobra) for its command-line interface. Cobra provides subcommands, flags, help text, and auto-generated man pages.

## Structure

The CLI entry point is `cmd/cli/main.go`. All command logic lives in `internal/cli/`:

```
internal/cli/
├── root.go       # Root command definition
├── migrate.go    # Migration subcommand
└── version.go    # Version subcommand
```

## Root Command

The root command is defined in `internal/cli/root.go`:

```go
var rootCmd = &cobra.Command{
    Use:   "{{ cookiecutter.project_slug }}",
    Short: "{{ cookiecutter.project_name }}",
    Long:  `{{ cookiecutter.project_description }}`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.AddCommand(migrateCmd)
    rootCmd.AddCommand(versionCmd)
}
```

The `init()` function registers all subcommands. Add new subcommands here.

## Subcommands

Each subcommand is a `*cobra.Command` defined in its own file:

```go
var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations using goose",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Command logic
        return nil
    },
}
```

### `RunE` vs `Run`

| Field | Return Type | When to Use |
|-------|-------------|-------------|
| `RunE` | `error` | When the command can fail (I/O, network, database) |
| `Run` | none | When the command cannot fail (printing help, version) |

Prefer `RunE` — it lets cobra handle errors cleanly and set the exit code.

## Flags

Add flags to a command with the `Flags()` methods:

```go
var configFile string

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations using goose",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Use configFile
        return nil
    },
}

func init() {
    migrateCmd.Flags().StringVarP(&configFile, "config", "c", "configs/config.yaml", "Config file path")
}
```

### Flag Types

| Method | Type | Example |
|--------|------|---------|
| `StringVarP` | string | `--config`, `-c` |
| `IntVarP` | int | `--port`, `-p` |
| `BoolVarP` | bool | `--verbose`, `-v` |
| `StringSliceVarP` | []string | `--tags`, `-t` |
| `DurationVarP` | time.Duration | `--timeout` |

The `P` suffix adds a shorthand flag (single dash). Omit it for long-only flags.

### Persistent Flags

Flags available on all subcommands:

```go
func init() {
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
```

### Local Flags

Flags available only on a specific subcommand:

```go
func init() {
    migrateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path")
}
```

## Adding a New Command

1. Create a new file in `internal/cli/` (e.g., `seed.go`).
2. Define the command:

   ```go
   var seedCmd = &cobra.Command{
       Use:   "seed",
       Short: "Seed the database with initial data",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Command logic
           return nil
       },
   }
   ```

3. Register it in `root.go`:

   ```go
   func init() {
       rootCmd.AddCommand(migrateCmd)
       rootCmd.AddCommand(versionCmd)
       rootCmd.AddCommand(seedCmd)
   }
   ```

4. Test it:

   ```bash
   just run-cli seed
   just run-cli seed --help
   ```

## Help Text Conventions

Follow these conventions for consistent help output:

| Field | Convention | Example |
|-------|------------|---------|
| `Use` | Lowercase, kebab-case for subcommands | `migrate`, `data-import` |
| `Short` | One line, sentence case, no period | `"Run database migrations"` |
| `Long` | Multi-line, detailed explanation | `"Run database migrations using goose. This applies all pending migrations to the configured database."` |
| `Example` | Indented code block with command | `"  {{ cookiecutter.project_slug }} migrate\n  {{ cookiecutter.project_slug }} migrate --config=custom.yaml"` |
| Flag `--help` | Describes what the flag does | `"Config file path"` |

### Example with All Fields

```go
var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations using goose",
    Long: `Run database migrations using goose.

This applies all pending migrations to the configured database.
Migrations are stored in the migrations/ directory and are
applied in order.`,
    Example: `  {{ cookiecutter.project_slug }} migrate
  {{ cookiecutter.project_slug }} migrate --config=configs/config.yaml`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
    },
}
```

## Man Page Generation

cobra can generate man pages from command definitions:

```go
err := rootCmd.GenManTree("/path/to/man")
```

This produces `.1` man files for the root command and each subcommand. Include this in your release process if you distribute the CLI as a standalone binary.

## Error Handling

Return errors from `RunE` — cobra prints them and exits with code 1:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    conn, err := service.NewDBConn(cfg.Database)
    if err != nil {
        return fmt.Errorf("connect database: %w", err)
    }
    defer conn.Close()

    db := service.NewBunDB(cfg.Database, conn)
    service.Init(db, slog.Default(), cfg.Auth.JwtSecret, cfg.Auth.APIKey)
    defer service.Close()

    return migrate.Run(db, slog.Default())
}
```

Wrap errors with context using `fmt.Errorf("context: %w", err)` so users know which step failed.

## Running the CLI

```bash
# Run the CLI
just run-cli

# Run a specific subcommand
just run-cli migrate

# View help
just run-cli --help
just run-cli migrate --help
```
