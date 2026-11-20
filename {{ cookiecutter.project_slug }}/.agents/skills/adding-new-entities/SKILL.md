---
name: adding-new-entities
description: "Use when adding a new domain entity (model, service, handler, migrations). Covers the full stack from database schema to API endpoints. Don't use for modifying existing entities."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Adding a new entity requires files in four packages plus migration files. Follow this order: model → service → handler → migrations.

## File structure

```
internal/
  model/
    post.go             # Domain struct with BaseModel embedding
  service/
    post.go             # CRUD functions using package-level db
  handler/
    post.go             # Huma endpoints, SetupPost(api, logger)
migrations/
  pg/
    002_create_posts.sql    # PostgreSQL migration
  sqlite/
    002_create_posts.sql    # SQLite migration
```

## Step 1: Model

Create `internal/model/{entity}.go`. Embed `BaseModel` for ID, timestamps, soft delete.

```go
package model

type Post struct {
    BaseModel `bson:",inline"`
    Title     string `bun:"type:text;not null"`
    Body      string `bun:"type:text"`
    Published bool   `bun:"type:bool;default:false"`
}
```

Use `bun:"type:..."` tags for non-default column types. Omit the tag for standard types.

## Step 2: Service

Create `internal/service/{entity}.go`. Plain functions, not methods. Use the package-level `db` variable.

```go
package service

import (
    "context"
    "database/sql"
    "time"

    "{{ cookiecutter.import_path }}/internal/model"

    "github.com/google/uuid"
)

// PostRead is the output schema for returning a post.
type PostRead struct {
    ID        uuid.UUID `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    Published bool      `json:"published"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func toPostRead(p model.Post) PostRead {
    return PostRead{
        ID:        p.ID,
        Title:     p.Title,
        Body:      p.Body,
        Published: p.Published,
        CreatedAt: p.CreatedAt,
        UpdatedAt: p.UpdatedAt,
    }
}

func GetPost(ctx context.Context, id uuid.UUID) (PostRead, error) {
    var p model.Post
    err := db.NewSelect().Model(&p).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return PostRead{}, ErrNotFound
        }
        return PostRead{}, err
    }
    return toPostRead(p), nil
}

func CreatePost(ctx context.Context, title, body string, published bool) (PostRead, error) {
    if title == "" {
        return PostRead{}, ErrInvalid
    }

    p := model.Post{
        BaseModel: BaseModel{ID: uuid.New()},
        Title:     title,
        Body:      body,
        Published: published,
    }
    if _, err := db.NewInsert().Model(&p).Exec(ctx); err != nil {
        return PostRead{}, err
    }
    return toPostRead(p), nil
}

func ListPosts(ctx context.Context, limit, offset int64) ([]PostRead, error) {
    var posts []model.Post
    if err := db.NewSelect().Model(&posts).Limit(limit).Offset(offset).OrderExpr("created_at DESC").Scan(ctx); err != nil {
        return nil, err
    }
    result := make([]PostRead, 0, len(posts))
    for _, p := range posts {
        result = append(result, toPostRead(p))
    }
    return result, nil
}

func ListPostsWithCount(ctx context.Context, limit, offset int64) ([]PostRead, int64, error) {
    posts, err := ListPosts(ctx, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    total, err := db.NewSelect().Model(&model.Post{}).Count(ctx)
    if err != nil {
        return nil, 0, err
    }
    return posts, int64(total), nil
}

func UpdatePost(ctx context.Context, id uuid.UUID, title, body *string, published *bool) (PostRead, error) {
    var p model.Post
    if err := db.NewSelect().Model(&p).Where("id = ?", id).Scan(ctx); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return PostRead{}, ErrNotFound
        }
        return PostRead{}, err
    }
    if title != nil {
        p.Title = *title
    }
    if body != nil {
        p.Body = *body
    }
    if published != nil {
        p.Published = *published
    }
    if _, err := db.NewUpdate().Model(&p).Where("id = ?", p.ID).Exec(ctx); err != nil {
        return PostRead{}, err
    }
    return toPostRead(p), nil
}

func DeletePost(ctx context.Context, id uuid.UUID) error {
    _, err := db.NewSelect().Model(&model.Post{}).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrNotFound
        }
        return err
    }
    _, err = db.NewDelete().Model(&model.Post{}).Where("id = ?", id).Exec(ctx)
    return err
}
```

**Key patterns:**
- `*string` / `*bool` pointers for optional update fields
- Check `sql.ErrNoRows` → return `ErrNotFound`
- `_, err := db.NewInsert().Model(&x).Exec(ctx)` — Exec returns 2 values
- `_, err := db.NewUpdate().Model(x).Where("id = ?", x.ID).Exec(ctx)` — same
- Conversion function (`toPostRead`) keeps output schema separate from model

### Service lifecycle

**Service package uses package-level variables.**

`db`, `logger`, `jwtSecret`, and `apiKey` are package-level variables initialized by `service.Init()`. The `service.Init(db, logger, secret, apiKey)` call happens in `main.go` before the server starts. `service.Close()` should be deferred in the `OnStart` hook (not the outer callback, which returns before the server starts).

Because these are global variables, tests cannot run in parallel (`-parallel` flag will cause flaky failures). Each test must call `service.Init()` with its own test database and `t.Cleanup(func() { _ = service.Close() })`.

## Step 3: Handler

Create `internal/handler/{entity}.go`. Takes `logger *slog.Logger`, not a service struct.

```go
package handler

import (
    "context"
    "errors"
    "log/slog"
    "net/http"

    "{{ cookiecutter.import_path }}/internal/service"

    "github.com/danielgtaylor/huma/v2"
    "github.com/google/uuid"
)

// PostListResponse wraps a paginated list with metadata.
type PostListResponse struct {
    Items  []service.PostRead `json:"items"`
    Total  int64              `json:"total"`
    Limit  int64              `json:"limit"`
    Offset int64              `json:"offset"`
}

// SetupPost registers all API routes for posts.
func SetupPost(api huma.API, logger *slog.Logger) {
    // GET /api/v1/posts
    huma.Get(api, "/api/v1/posts", func(ctx context.Context, input *struct {
        Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20"`
        Offset int `query:"offset" minimum:"0"`
    }) (*struct{ Body PostListResponse }, error) {
        items, total, err := service.ListPostsWithCount(ctx, int64(input.Limit), int64(input.Offset))
        if err != nil {
            logger.ErrorContext(ctx, "list posts failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body PostListResponse }{
            Body: PostListResponse{
                Items:  items,
                Total:  total,
                Limit:  int64(input.Limit),
                Offset: int64(input.Offset),
            },
        }, nil
    })

    // POST /api/v1/posts
    huma.Post(api, "/api/v1/posts", func(ctx context.Context, input *struct {
        Body struct {
            Title     string `json:"title" maxLength:"255"`
            Body      string `json:"body" maxLength:"100000"`
            Published bool   `json:"published"`
        }
    }) (*struct {
        Status int
        Body   service.PostRead
    }, error) {
        post, err := service.CreatePost(ctx, input.Body.Title, input.Body.Body, input.Body.Published)
        if err != nil {
            if errors.Is(err, service.ErrInvalid) {
                return nil, huma.NewError(http.StatusBadRequest, "invalid input")
            }
            logger.ErrorContext(ctx, "create post failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct {
            Status int
            Body   service.PostRead
        }{Status: http.StatusCreated, Body: post}, nil
    })

    // GET /api/v1/posts/{id}
    huma.Get(api, "/api/v1/posts/{id}", func(ctx context.Context, input *struct {
        ID uuid.UUID `path:"id"`
    }) (*struct{ Body service.PostRead }, error) {
        post, err := service.GetPost(ctx, input.ID)
        if err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "post not found")
            }
            logger.ErrorContext(ctx, "get post failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body service.PostRead }{Body: post}, nil
    })

    // PUT /api/v1/posts/{id}
    huma.Put(api, "/api/v1/posts/{id}", func(ctx context.Context, input *struct {
        ID uuid.UUID `path:"id"`
        Body struct {
            Title     *string `json:"title" maxLength:"255"`
            Body      *string `json:"body" maxLength:"100000"`
            Published *bool   `json:"published"`
        }
    }) (*struct{ Body service.PostRead }, error) {
        post, err := service.UpdatePost(ctx, input.ID, input.Body.Title, input.Body.Body, input.Body.Published)
        if err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "post not found")
            }
            logger.ErrorContext(ctx, "update post failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body service.PostRead }{Body: post}, nil
    })

    // DELETE /api/v1/posts/{id}
    huma.Delete(api, "/api/v1/posts/{id}", func(ctx context.Context, input *struct {
        ID uuid.UUID `path:"id"`
    }) (*struct{ Status int }, error) {
        if err := service.DeletePost(ctx, input.ID); err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "post not found")
            }
            logger.ErrorContext(ctx, "delete post failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Status int }{Status: http.StatusNoContent}, nil
    })
}
```

**Key patterns:**
- Always log the real error: `logger.ErrorContext(ctx, "action failed", "error", err)`
- Return generic message to client: `"internal server error"`
- Check sentinel errors: `errors.Is(err, service.ErrNotFound)` → 404, `ErrInvalid` → 400
- PUT uses `*string` / `*bool` for optional fields
- DELETE returns `http.StatusNoContent`

## Step 4: Register the handler

In `internal/handler/items.go`, add to the `Setup` function:

```go
func Setup(api huma.API, logger *slog.Logger) {
    SetupAuth(api, logger)
    // ... existing routes ...
    SetupPost(api, logger)  // add this line
}
```

## Step 5: Migrations

Create matching files in both `migrations/pg/` and `migrations/sqlite/`:

```sql
-- migrations/sqlite/002_create_posts.sql
-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT DEFAULT '',
    published BOOLEAN DEFAULT FALSE,
    created_by UUID,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

-- +goose Down
DROP TABLE IF EXISTS posts;
```

Use `just migrate-create create_posts` to generate the numbered files, then edit the SQL.

**Type mapping:**

| Go type | SQLite | PostgreSQL |
|---------|--------|------------|
| `uuid.UUID` | `UUID` | `UUID` |
| `string` | `TEXT` | `TEXT` |
| `int` | `INTEGER` | `INTEGER` |
| `bool` | `BOOLEAN` | `BOOLEAN` |
| `time.Time` | `DATETIME` | `TIMESTAMPTZ` |
| `float64` | `REAL` | `DOUBLE PRECISION` |

## Key conventions

- **No repository layer.** Queries live in the service package using package-level `db`.
- **Plain functions, not methods.** `func GetPost(ctx, id)` not `func (s *Service) GetPost()`.
- **Service takes no struct.** Uses package-level `db`, `logger`, `jwtSecret`.
- **Handler takes logger.** `SetupX(api, logger)` — never takes a service struct.
- **Error logging.** Always `logger.ErrorContext(ctx, "msg", "error", err)` before returning generic error.
- **Sentinel errors.** `ErrNotFound` and `ErrInvalid` from the service package.
- **Migration numbering.** Check existing migrations and use the next number. Both pg and sqlite get the same number.
