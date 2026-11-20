---
name: adding-api-endpoints
description: "Use when adding, modifying, or debugging Huma v2 API endpoints. Don't use for CLI commands, background tasks, or database queries."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+, Huma v2, Chi v5"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

Add new endpoints in `internal/handler/` using `huma.Get`, `huma.Post`, `huma.Put`, `huma.Delete`. Register them in the `Setup()` function. Handlers call service methods — never touch the database directly.

## Workflow

1. Define input and output structs in `internal/handler/` (or reuse from `internal/model/`)
2. Register the handler in `Setup()` in `internal/handler/`
3. Call the appropriate service method from the handler
4. Use `huma.NewError()` for error responses (v2 API)

### Full CRUD example

```go
// internal/handler/items.go
package handler

import (
    "context"
    "net/http"

    "{{ cookiecutter.import_path }}/internal/service"

    "github.com/danielgtaylor/huma/v2"
    "github.com/google/uuid"
)

func Setup(api huma.API, logger *slog.Logger) {
    // GET /api/v1/items — list with pagination
    huma.Get(api, "/api/v1/items", func(ctx context.Context, input *struct {
        Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Max items"`
        Offset int `query:"offset" minimum:"0" doc:"Offset"`
    }) (*struct{ Body ListResponse }, error) {
        items, total, err := service.ListItemsWithCount(ctx, int64(input.Limit), int64(input.Offset))
        if err != nil {
            logger.ErrorContext(ctx, "list items failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body ListResponse }{
            Body: ListResponse{
                Items:  items,
                Total:  total,
                Limit:  int64(input.Limit),
                Offset: int64(input.Offset),
            },
        }, nil
    })

    // POST /api/v1/items — create
    huma.Post(api, "/api/v1/items", func(ctx context.Context, input *struct {
        Body struct {
            Name        string `json:"name" maxLength:"255" doc:"Item name"`
            Description string `json:"description" maxLength:"10000" doc:"Item description"`
        }
    }) (*struct {
        Status int
        Body   service.ItemRead
    }, error) {
        item, err := service.CreateItem(ctx, input.Body.Name, input.Body.Description)
        if err != nil {
            logger.ErrorContext(ctx, "create item failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct {
            Status int
            Body   service.ItemRead
        }{Status: http.StatusCreated, Body: item}, nil
    })

    // GET /api/v1/items/{id} — get by ID
    huma.Get(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
        ID uuid.UUID `path:"id" doc:"Item ID"`
    }) (*struct{ Body service.ItemRead }, error) {
        item, err := service.GetItem(ctx, input.ID)
        if err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "item not found")
            }
            logger.ErrorContext(ctx, "get item failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body service.ItemRead }{Body: item}, nil
    })

    // PUT /api/v1/items/{id} — partial update (pointer fields)
    huma.Put(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
        ID   uuid.UUID `path:"id" doc:"Item ID"`
        Body struct {
            Name        *string `json:"name" maxLength:"255" doc:"Item name"`
            Description *string `json:"description" maxLength:"10000" doc:"Item description"`
        }
    }) (*struct{ Body service.ItemRead }, error) {
        item, err := service.UpdateItem(ctx, input.ID, input.Body.Name, input.Body.Description)
        if err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "item not found")
            }
            logger.ErrorContext(ctx, "update item failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Body service.ItemRead }{Body: item}, nil
    })

    // DELETE /api/v1/items/{id} — soft delete
    huma.Delete(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
        ID uuid.UUID `path:"id" doc:"Item ID"`
    }) (*struct{ Status int }, error) {
        if err := service.DeleteItem(ctx, input.ID); err != nil {
            if errors.Is(err, service.ErrNotFound) {
                return nil, huma.NewError(http.StatusNotFound, "item not found")
            }
            logger.ErrorContext(ctx, "delete item failed", "error", err)
            return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
        }
        return &struct{ Status int }{Status: http.StatusNoContent}, nil
    })
}
```

### Input struct patterns

**Query parameters** — use value types with `default:` tags. Pointers cause panics in Huma v2:

```go
// GOOD — value type with default
Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20"`

// BAD — pointer causes "pointers are not supported" panic
Limit  *int `query:"limit"`
```

**JSON body fields** — use pointer types for optional fields:

```go
// Required field (non-pointer)
Name string `json:"name" maxLength:"255" doc:"Item name"`

// Optional field (pointer — omitted in JSON means nil)
Description *string `json:"description" maxLength:"10000" doc:"Item description"`
```

**Path parameters** — typed extraction:

```go
ID uuid.UUID `path:"id" doc:"Item ID"`
// Also supports: int, string, bool
```

### Output status codes

```go
// 200 OK (default — just return Body)
return &struct{ Body service.ItemRead }{Body: item}, nil

// 201 Created
return &struct {
    Status http.StatusCreated
    Body   service.ItemRead
}{Status: http.StatusCreated, Body: item}, nil

// 204 No Content
return &struct{ Status int }{Status: http.StatusNoContent}, nil
```

### Error responses

Use `huma.NewError(code, message, err)` — this is the v2 API. The old `huma.Error500Internal()` does not exist in v2:

```go
// GOOD (v2)
return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
return nil, huma.NewError(http.StatusNotFound, "not found")
return nil, huma.NewError(http.StatusBadRequest, "bad request")

// WRONG (v1 style, doesn't exist in v2)
return nil, huma.Error500Internal("message", err)
```

Validation failures from struct tags automatically return 422 with detailed error info — no manual handling needed.

## Gotchas

**FieldsOptionalByDefault: true**

This project's Huma config sets `FieldsOptionalByDefault: true`. Clients can omit any body field. Use pointer types only to distinguish "client sent a zero value" from "client omitted the field":

```go
// Optional — client can omit entirely
Name string `json:"name"`

// Optional — use pointer to distinguish false from omitted
Enabled *bool `json:"enabled" default:"true"`
```

**Pagination requires named types**

You cannot return an inline struct literal from a handler whose signature uses an inline struct. Define a named type:

```go
// Define at package level
type ListResponse struct {
    Items []service.ItemRead `json:"items"`
    Total int64              `json:"total"`
}

// Use in the handler
}) (*struct {
    Body ListResponse
}, error) {
    return &struct{ Body ListResponse }{
        Body: ListResponse{Items: items, Total: total},
    }, nil
}
```

**Auth: Adding public routes**

Default public paths are `/health`, `/metrics`, `/static/`, `/openapi.json`, `/openapi.yaml`, `/api/v1/auth/login`. Add more via `AuthOpts` with `PublicPaths` (all methods) and `PublicReadPaths` (GET/HEAD/OPTIONS only):

```go
r.Use(middleware.Auth(middleware.AuthOpts{
    Secret: cfg.Auth.JwtSecret,
    // PublicPaths — bypass auth for ALL methods on these paths
    PublicPaths: []string{"/pages", "/pages/"},
    // PublicReadPaths — bypass auth for GET/HEAD/OPTIONS only.
    // POST/PUT/PATCH/DELETE still require authentication.
    PublicReadPaths: []string{"/api/v1/pages", "/api/v1/categories"},
}))
```

**Web routes: Use `*chi.Mux`, not `http.ServeMux`**

Chi does not delegate to embedded standard-library muxes. Register web routes directly on the Chi router:

```go
func SetupWeb(r *chi.Mux, ren *renderer.Renderer) {
    r.Get("/", func(w http.ResponseWriter, req *http.Request) {
        ren.Render(w, "home.html", map[string]any{"Title": "Home"})
    })
    r.Get("/pages/{slug}", func(w http.ResponseWriter, req *http.Request) {
        slug := chi.URLParam(req, "slug")
        // ...
    })
}
```

**Huma v2: `$schema` field in all responses**

Huma v2 adds a `$schema` field to every JSON response pointing to the OpenAPI schema URL. To disable: set `config.SchemaProvider.TypeSchemaGetter = nil` in the Huma config.

## Key conventions

- Separate input/output structs — never reuse the same struct for both
- Handlers call service methods, never touch the database directly
- Always pass `context.Context` as the first handler parameter
- Use `json:"-"` to exclude fields from API responses (e.g., `DeletedAt`)
- Add `doc:` tags to all struct fields for OpenAPI documentation
