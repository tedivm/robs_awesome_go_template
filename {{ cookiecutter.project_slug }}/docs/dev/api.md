# REST API — Huma + Chi

{{ cookiecutter.project_name }} uses [Huma v2](https://github.com/danielgtaylor/huma) as its OpenAPI-first API layer, with [Chi](https://github.com/go-chi/chi) as the router. Huma generates OpenAPI 3.1 documentation, validates requests, and produces RFC 9457 error responses — all automatically from Go struct tags.

## Struct-Based Endpoints

Endpoints are defined as Go functions with typed input and output structs. Huma inspects the struct definitions to generate OpenAPI schemas, validate incoming requests, and serialize responses.

```go
huma.Get(api, "/api/v1/items", func(ctx context.Context, input *struct {
    Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Max items to return"`
    Offset int `query:"offset" minimum:"0" default:"0" doc:"Offset for pagination"`
}) (*struct {
    Body []service.ItemRead
}, error) {
    // Handler logic
})
```

Input structs define path parameters, query parameters, and request bodies. Output structs define the response body and HTTP status code.

## Tag-Based Validation

Huma uses struct tags for validation, following JSON Schema conventions. Validation runs automatically before the handler executes.

| Tag | Description | Example |
|-----|-------------|---------|
| `maxLength` | Maximum string length | `maxLength:"255"` |
| `minLength` | Minimum string length | `minLength:"1"` |
| `maximum` | Maximum numeric value | `maximum:"100"` |
| `minimum` | Minimum numeric value | `minimum:"0"` |
| `pattern` | Regex pattern | `pattern:"^[a-z]+$"` |
| `enum` | Allowed values | `enum:"active,inactive"` |
| `required` | Required field | `required:"true"` |
| `doc` | Human-readable description | `doc:"Item name"` |

```go
type ItemCreate struct {
    Name        string `json:"name" maxLength:"255" doc:"Item name"`
    Description string `json:"description" maxLength:"10000" doc:"Item description"`
}
```

If validation fails, Huma returns a 422 Unprocessable Entity response with an RFC 9457 `application/problem+json` body listing each validation error.

## Automatic OpenAPI Documentation

Huma serves the OpenAPI 3.1 specification automatically:

| Path | Format | Description |
|------|--------|-------------|
| `/openapi.json` | JSON | OpenAPI 3.1 specification |
| `/openapi.yaml` | YAML | OpenAPI 3.1 specification |

The spec is generated at runtime from the struct definitions — no separate documentation file to maintain. For an interactive docs UI, integrate [Stoplight Elements](https://github.com/stoplightio/elements) or [Scalar](https://github.com/scalar/scalar) by serving a static HTML page that points to `/openapi.json`.

## RFC 9457 Error Responses

Huma produces standardized error responses following [RFC 9457 (Problem Details for HTTP APIs)](https://datatracker.ietf.org/doc/html/rfc9457). All error responses use the `application/problem+json` content type.

| Status | Huma Function | When to Use |
|--------|---------------|-------------|
| 400 | `huma.NewError(400, "Bad Request", err)` | Bad request (general) |
| 401 | `huma.NewError(401, "Unauthorized", err)` | Missing or invalid authentication |
| 403 | `huma.NewError(403, "Forbidden", err)` | Valid credentials, insufficient permissions |
| 404 | `huma.NewError(404, "Not Found", err)` | Resource not found |
| 409 | `huma.NewError(409, "Conflict", err)` | Resource conflict (e.g., duplicate) |
| 422 | Auto-generated | Validation failures from struct tags |
| 500 | `huma.NewError(500, "Internal Server Error", err)` | Unexpected server error |

Example error response:

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "item not found",
  "instance": "/api/v1/items/abc123"
}
```

## Input/Output Struct Separation

The template separates input and output schemas to control what data enters and leaves the API:

| Struct | Purpose | Location |
|--------|---------|----------|
| Input structs | Input for POST/PUT | Inline in `internal/handler/items.go` |
| `ItemRead` | Output for GET | `internal/service/item.go` |
| `Item` | Database model | `internal/model/item.go` |

This pattern prevents leaking internal fields (like `DeletedAt`) in API responses and allows different validation rules for creation vs. update.

## Path, Query, and Body Parameters

### Path Parameters

Use the `path` tag to extract typed path parameters:

```go
input *struct {
    ID uuid.UUID `path:"id" doc:"Item ID"`
}
```

Huma automatically parses and validates the type. If the client sends an invalid UUID, Huma returns a 422 error before the handler runs.

### Query Parameters

Use the `query` tag with value types and `default:` tags for query parameters with sensible defaults:

```go
input *struct {
    Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Max items"`
    Offset int `query:"offset" minimum:"0" default:"0" doc:"Offset"`
}
```

Huma v2 does not support pointer types for query parameters — use value types with `default:` tags instead.

### Request Body

Nest the body in a `Body` struct within the input:

```go
input *struct {
    Body struct {
        Name        string `json:"name" maxLength:"255" doc:"Item name"`
        Description string `json:"description" maxLength:"10000" doc:"Item description"`
    }
}
```

## Chi Middleware

Chi provides production-grade middleware for the HTTP layer. The template configures these middleware in `main.go`:

| Middleware | Purpose |
|------------|---------|
| `middleware.Recoverer` | Catch panics and return 500 responses |
| `middleware.RequestID` | Unique request ID for tracing |
| `middleware.RequestLogger` | Structured request logging via `log/slog` |
| `middleware.RealIP` | Extract real client IP from proxies |
| `middleware.StripSlashes` | Remove trailing slashes from URLs |

Add middleware with `r.Use()`:

```go
r := chi.NewMux()
r.Use(middleware.Recoverer)
r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
```

## Adding a New Endpoint

1. Define input and output structs in `internal/handler/` with appropriate struct tags.
2. Register the endpoint with `huma.Get()`, `huma.Post()`, `huma.Put()`, or `huma.Delete()`.
3. Implement the handler logic, delegating to the service layer.
4. Return the output struct or a Huma error function.
5. The OpenAPI docs update automatically — no extra work needed.
