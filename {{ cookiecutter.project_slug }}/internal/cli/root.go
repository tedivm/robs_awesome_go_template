package cli

import (
	"log/slog"

	"{{ cookiecutter.import_path }}/internal/version"

	"github.com/spf13/cobra"
)

// NewRootCommand creates the root CLI command.
func NewRootCommand(logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "{{ cookiecutter.project_slug }}",
		Short:   "{{ cookiecutter.project_name }}",
		Long:    "{{ cookiecutter.project_description }}",
		Version: version.Version,
	}

	cmd.SetVersionTemplate("{{ cookiecutter.project_slug }} v{% raw %}{{ .Version }}{% endraw %}\n")

	// Add subcommands
	cmd.AddCommand(newMigrateCommand(logger))
	cmd.AddCommand(newVersionCommand())

	return cmd
}
