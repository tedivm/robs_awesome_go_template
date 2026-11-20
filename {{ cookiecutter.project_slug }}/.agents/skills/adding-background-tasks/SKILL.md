---
name: adding-background-tasks
description: "Use when adding background tasks with asynq (distributed, I/O-bound) or Go worker pools (CPU-bound). Don't use for synchronous API handlers or CLI commands."
license: MIT
compatibility: "Go {{ cookiecutter.go_version }}+, asynq, Redis"
metadata:
  author: Robert Hafner
  source: "https://github.com/tedivm/robs_awesome_go_template"
---

## Quick start

For distributed I/O-bound tasks: define task type constants (`namespace:action` format), payload structs, builder functions, and handlers in `internal/worker/`. For CPU-bound tasks: use the Go worker pool in `internal/worker/pool.go`.

## Workflow

### Adding an asynq task

**1. Define a task type constant** — use `namespace:action` format:

```go
// internal/worker/tasks.go
package worker

const (
    TaskTypeProcessData  = "data:process"
    TaskTypeSendEmail    = "email:send"
    TaskTypeGenerateReport = "report:generate"
)
```

**2. Define a payload struct** — keep it small and serializable:

```go
type SendEmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

**3. Create a task builder function:**

```go
func NewSendEmailTask(to, subject, body string) *asynq.Task {
    payload, _ := json.Marshal(SendEmailPayload{
        To:      to,
        Subject: subject,
        Body:    body,
    })
    return asynq.NewTask(TaskTypeSendEmail, payload,
        asynq.MaxRetry(3),
        asynq.Queue("default"),
        asynq.ProcessWithin(10*time.Minute),
    )
}
```

Common task options:
- `asynq.MaxRetry(N)` — maximum retry attempts
- `asynq.Queue("name")` — queue name (`critical`, `default`, `low`)
- `asynq.ProcessWithin(d)` — deadline for processing
- `asynq.TaskID(id)` — custom task ID for deduplication

**4. Create a handler** implementing `asynq.Handler`:

```go
func SendEmailHandler(logger *slog.Logger) asynq.Handler {
    return &sendEmailHandler{logger: logger}
}

type sendEmailHandler struct {
    logger *slog.Logger
}

func (h *sendEmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
    var payload SendEmailPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    h.logger.Info("sending email",
        slog.String("to", payload.To),
        slog.String("subject", payload.Subject),
    )

    // Call service package functions directly (no struct)
    if err := service.SendEmail(ctx, payload.To, payload.Subject, payload.Body); err != nil {
        return fmt.Errorf("send email: %w", err)
    }
    return nil
}
```

**5. Register the handler with the server:**

```go
// In the server setup (e.g., main.go or worker package)
mux := asynq.NewServeMux()
mux.HandleFunc(TaskTypeProcessData, ProcessDataHandler(logger, svc))
mux.HandleFunc(TaskTypeSendEmail, SendEmailHandler(logger, svc))
mux.HandleFunc(TaskTypeGenerateReport, GenerateReportHandler(logger, svc))

server := asynq.NewServer(redisConn, mux,
    asynq.Concurrency(10),
    asynq.Queues(map[string]int{
        "critical": 5,
        "default":  10,
        "low":      5,
    }),
)
```

**Note:** `RedisConnOpt` is an interface value, not a pointer to interface:

```go
// CORRECT — pass the interface value
func NewServer(redisConn asynq.RedisConnOpt) *asynq.Server

// WRONG — pointer to interface causes "pointer to interface" error
func NewServer(redisConn *asynq.RedisConnOpt) *asynq.Server
```

**6. Enqueue a task from a handler or service:**

```go
task := worker.NewSendEmailTask("user@example.com", "Welcome", "Hello!")
if err := producer.Enqueue(task); err != nil {
    logger.Error("failed to enqueue task", slog.String("error", err.Error()))
    return err
}
```

### Periodic tasks (Scheduler)

```go
scheduler := asynq.NewScheduler(redisConn, asynq.SchedulerOpts{
    GenID: func(t *asynq.Task) string { return "periodic-cleanup" },
})

entry := asynq.NewTask("cleanup:run", nil)
cronExpr := "0 * * * *" // Every hour
err := scheduler.Register(cronExpr, entry)
```

### CPU-bound tasks (Go worker pool)

For tasks that don't need persistence or distributed execution:

```go
pool := worker.NewPool(ctx, 4, logger)
pool.Start()

// Submit tasks
pool.Submit(func(ctx context.Context) error {
    // CPU-bound work here
    return processItem(ctx, itemID)
})

// Graceful shutdown
pool.Stop()
```

## Key conventions

- Task types use `namespace:action` format: `data:process`, `email:send`
- Payloads are small JSON structs — don't pass large data or database connections
- Handlers return errors for retry — asynq retries automatically up to `MaxRetry`
- Use queues for priority: `critical` (5 workers), `default` (10), `low` (5)
- Tasks live in `internal/worker/` — one file per task type or grouped logically
- Redis is required for asynq — use Go worker pools for CPU-bound tasks without Redis
- `RedisConnOpt` is an interface value, not a pointer to interface
