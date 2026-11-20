package cli

import (
	"fmt"

	"{{ cookiecutter.import_path }}/internal/version"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("{{ cookiecutter.project_slug }} v%s\n", version.Version)
			fmt.Printf("commit: %s\n", version.Commit)
			fmt.Printf("built:  %s\n", version.Date)
		},
	}
}
