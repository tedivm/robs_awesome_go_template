package main

import (
	"log/slog"
	"os"

	"{{ cookiecutter.import_path }}/internal/cli"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	rootCmd := cli.NewRootCommand(logger)

	if err := rootCmd.Execute(); err != nil {
		logger.Error("cli error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
