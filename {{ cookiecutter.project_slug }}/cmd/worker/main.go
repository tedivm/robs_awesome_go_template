package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"{{ cookiecutter.import_path }}/internal/config"
	"{{ cookiecutter.import_path }}/internal/worker"

	"github.com/hibiken/asynq"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	// Create logger
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

	if cfg.Redis.Addr == "" {
		logger.Error("redis addr is not set — cannot start worker")
		os.Exit(1)
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	srv := worker.NewServer(redisOpt, logger)
	mux := worker.NewMux(logger)

	logger.Info("worker starting")

	go func() {
		if err := srv.Run(mux); err != nil {
			logger.Error("worker server failed", slog.String("error", err.Error()))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down worker")
	srv.Shutdown()
	logger.Info("worker stopped")
}
