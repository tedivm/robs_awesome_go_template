# Migrations — goose

{{ cookiecutter.project_name }} uses [goose](https://github.com/pressly/goose) for versioned database migrations. goose manages SQL migration files, tracks applied migrations in a `goose_db_version` table, and supports forward and backward migration.

## Workflow

### Create a New Migration

With CLI enabled:

```bash
just migrate-create <name>
```

This creates a single file in both `migrations/pg/` and `migrations/sqlite/`:

```
migrations/
├── pg/
│   ├── 001_init.sql
│   └── 002_add_items.sql
└── sqlite/
    ├── 001_init.sql
    └── 002_add_items.sql
```

Each file contains both `Up` and `Down` sections marked with `-- +goose Up` / `-- +goose Down` annotations.

### Apply Migrations

```bash
# Apply all pending migrations (via CLI)
just migrate-apply

# Check migration status
just migrate-plan
```

### Rollback Migrations

```bash
# Rollback the last migration (via CLI)
go run ./cmd/cli migrate down
```

## SQL File Format

goose v3 uses a **single file** per migration with both `Up` and `Down` sections:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
```

### Statement Boundaries

`+goose StatementBegin` and `+goose StatementEnd` wrap a single transactional unit. goose wraps each statement block in a transaction. Use multiple blocks for migrations with several operations:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items ADD COLUMN category_id UUID REFERENCES categories(id);
-- +goose StatementEnd
```

### Single Statement (No Wrapping Needed)

For a single SQL statement, you can omit `StatementBegin`/`StatementEnd`:

```sql
-- +goose Up
CREATE TABLE items (...);
```

## Environment Variable Substitution

goose supports environment variable substitution in SQL files using `${VAR_NAME}` syntax:

```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "${GOOSE_EXTENSION_SCHEMA}"."uuid-ossp";
```

Set the environment variable before running goose:

```bash
GOOSE_EXTENSION_SCHEMA=public goose -dir migrations up
```

This is useful for database-specific configurations like schema names or extension paths.

## Keeping Go Models and Migrations in Sync

bun models and goose migrations describe the same schema from different perspectives:

| Aspect | bun Model | goose Migration |
|--------|-----------|-----------------|
| Purpose | Runtime queries, type safety | Versioned schema changes |
| Source of truth | Business logic | Database schema |
| When to change | When adding fields to Go code | When the database schema changes |

### Workflow for Adding a New Table

1. **Create the migration first:**

   ```bash
   just migrate-create add_items
   ```

2. **Write the SQL in the migration file (single file, both Up and Down):**

   ```sql
   -- +goose Up
   CREATE TABLE items (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(255) NOT NULL,
       description TEXT,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       deleted_at TIMESTAMPTZ
   );

   -- +goose Down
   DROP TABLE IF EXISTS items;
   ```

3. **Write the Go model to match:**

   ```go
   type Item struct {
       BaseModel `bun:"table:items,alias:i"`
       Name        string `bun:"type:varchar(255),notnull" json:"name"`
       Description string `bun:"type:text" json:"description"`
   }
   ```

4. **Apply the migration:**

   ```bash
   just migrate-apply
   ```

### Rules

- **Always create a migration before changing the database schema.** Never rely on bun's auto-migration in production.
- **The migration is the source of truth for the database schema.** The Go model should match the migration, not the other way around.
- **bun's built-in migrator is for development only.** It auto-creates tables from model definitions. Use it for local development convenience, but always write goose migrations for production.
- **Test migrations against both SQLite and PostgreSQL.** SQLite has different type support. Use `TEXT` for strings in SQLite migrations if you need cross-database compatibility.

## Migration Execution Contexts

Migrations run in these contexts:

| Context | How |
|---------|-----|
| Local development | `just migrate-apply` |
| Docker Compose | CLI runs `go run ./cmd/cli migrate up` before the app starts |
| CLI command | `go run ./cmd/cli migrate up` |
| CI/CD | Production deployment pipeline |

Each context uses the same migration files but may connect to different databases (SQLite locally, PostgreSQL in production).
