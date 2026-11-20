package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"{{ cookiecutter.import_path }}/internal/config"
	"{{ cookiecutter.import_path }}/internal/handler"
	"{{ cookiecutter.import_path }}/internal/middleware"
{% if cookiecutter.include_orm == "y" %}
	"{{ cookiecutter.import_path }}/internal/service"
{% endif %}
{% if cookiecutter.include_workers == "y" %}
	"{{ cookiecutter.import_path }}/internal/worker"
{% endif %}

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
{% if cookiecutter.include_workers == "y" %}
	"github.com/hibiken/asynq"
{% endif %}

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// slogLogger adapts slog.Logger to chi's LoggerInterface.
type slogLogger struct {
	*slog.Logger
}

func (l *slogLogger) Print(v ...any) {
	l.Info(fmt.Sprint(v...))
}

// noDirListing wraps an http.Handler to reject directory listing requests.
// When a directory is requested (path ends with / or has no extension), it returns 404.
func noDirListing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the path ends with /, it's a directory request — reject it.
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Options defines CLI options for the server.
type Options struct {
	Port int `doc:"Port to listen on" short:"p" default:"8080"`
}

func main() {
	// Load config early to get log settings
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate config before use
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

{% if cookiecutter.include_web == "y" %}
	// Extra validation for production deployments
	if cfg.Server.Environment == "production" {
		if err := cfg.ValidateForProduction(); err != nil {
			fmt.Fprintf(os.Stderr, "production config invalid: %v\n", err)
			os.Exit(1)
		}
	}

{% endif %}

	// Create logger with configured level
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = slog.LevelInfo
	}

	var logHandler slog.Handler
	if cfg.Log.Format == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {

		// Static files — resolve early so renderer and other components can use it.
		// Check STATIC_PATH env var first (Docker dev), then CWD-relative (local dev),
		// then executable-relative (production).
		staticDir := os.Getenv("STATIC_PATH")
		if staticDir == "" {
			if _, err := os.Stat("static"); err == nil {
				staticDir = "static"
			} else {
				ex, _ := os.Executable()
				staticDir = filepath.Join(filepath.Dir(ex), "static")
			}
		}

		// Chi router with middleware
		r := chi.NewMux()
		r.Use(chimw.Recoverer)
		r.Use(chimw.RequestID)
		// Propagate request ID to response header
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if id := chimw.GetReqID(r.Context()); id != "" {
					w.Header().Set("X-Request-ID", id)
				}
				h.ServeHTTP(w, r)
			})
		})
		r.Use(chimw.RequestLogger(&chimw.DefaultLogFormatter{
			Logger: &slogLogger{logger},
		}))
		r.Use(chimw.StripSlashes)
		r.Use(middleware.MetricsMiddleware())

		// Security headers
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("Content-Security-Policy", "default-src 'self'")
				h.ServeHTTP(w, r)
			})
		})

		// Request context timeout
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
				defer cancel()
				r = r.WithContext(ctx)
				h.ServeHTTP(w, r)
			})
		})

		// CORS — handles preflight requests early
		r.Use(middleware.CORS(middleware.CORSCfg{
			AllowedOrigins: cfg.Cors.AllowedOrigins,
			AllowedMethods: cfg.Cors.AllowedMethods,
			AllowedHeaders: cfg.Cors.AllowedHeaders,
		}))

		// Rate limiting
		r.Use(chimw.ThrottleWithOpts(chimw.ThrottleOpts{
			Limit:          cfg.Server.RateLimit,
			BacklogTimeout: 60 * time.Second,
		}))

		// Authentication — skip if auth_method is "none"
		if cfg.Auth.AuthMethod != "none" && cfg.Auth.JwtSecret == "" {
			logger.Error("auth is enabled but jwt_secret is not set — refusing to start")
			os.Exit(1)
		}
		if cfg.Auth.AuthMethod == "api_key" && cfg.Auth.APIKey == "" {
			logger.Error("auth_method is api_key but api_key is not set — refusing to start")
			os.Exit(1)
		}
		r.Use(middleware.Auth(middleware.AuthOpts{Secret: cfg.Auth.JwtSecret}))

		// Wrap with Huma — FieldsOptionalByDefault makes body fields truly optional
		// (clients can omit them instead of sending null). Stable since v2.39.0.
		config := huma.DefaultConfig("{{ cookiecutter.project_name }}", "1.0.0")
		config.FieldsOptionalByDefault = true
		api := humachi.New(r, config)

{% if cookiecutter.include_orm == "y" %}
		// Initialize database connection
		conn, err := service.NewDBConn(cfg.Database)
		if err != nil {
			logger.Error("failed to open database", slog.String("error", err.Error()))
			os.Exit(1)
		}

		db := service.NewBunDB(cfg.Database, conn)
		service.Init(db, logger, cfg.Auth.JwtSecret, cfg.Auth.APIKey)

		switch cfg.Database.Driver {
		case "postgres":
			logger.Info("connected to postgres database")
		case "sqlite":
			logger.Info("connected to sqlite database")
		}

{% endif %}		// Setup routes
		handler.Setup(api, logger)

		// Health check (outside Huma for monitoring)
{% if cookiecutter.include_orm == "y" %}
		r.Get("/health", handler.HealthCheck(db, logger))
{% else %}
		r.Get("/health", handler.HealthCheck(logger))
{% endif %}

		// Metrics endpoint
		r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
			middleware.MetricsHandler().ServeHTTP(w, r)
		})

		// Static files — serve individual files only, no directory listings
		staticFS := http.Dir(staticDir)
		r.Handle("/static/*", http.StripPrefix("/static/", noDirListing(http.FileServer(staticFS))))

		// OpenAPI spec endpoints (automatic from Huma)
		// /openapi.json and /openapi.yaml are served automatically

		hooks.OnStart(func() {
{% if cookiecutter.include_orm == "y" %}
			// Close DB and connection on shutdown — must be here, not in the outer
			// callback, which returns before the server starts.
			defer service.Close()
			defer conn.Close()
{% endif %}
			srv := &http.Server{
				Addr:         fmt.Sprintf(":%d", options.Port),
				Handler:      r,
				ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
				WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
				IdleTimeout:  120 * time.Second,
			}

			// Start background workers
{% if cookiecutter.include_workers == "y" %}
			if cfg.Redis.Addr != "" {
				redisOpt := asynq.RedisClientOpt{
					Addr:     cfg.Redis.Addr,
					Password: cfg.Redis.Password,
					DB:       cfg.Redis.DB,
				}
				asynqSrv := worker.NewServer(redisOpt, logger)
				go func() {
					if err := asynqSrv.Run(worker.NewMux(logger)); err != nil {
						logger.Error("worker server failed", slog.String("error", err.Error()))
					}
				}()
				logger.Info("worker server started")
			}
{% endif %}

			go func() {
				logger.Info("server starting", slog.String("addr", srv.Addr))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("server failed", slog.String("error", err.Error()))
				}
			}()

			// Graceful shutdown
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			logger.Info("shutting down server")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			srv.Shutdown(ctx)
		})
	})

	cli.Run()
}
