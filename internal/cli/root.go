package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newRootCommand(stdout io.Writer, stderr io.Writer, application Application) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:           "skill-switcher",
		Short:         "Manage shared skills across supported coding agents",
		Long:          "skill-switcher manages a shared skill catalog, refreshes upstream sources, and syncs selected skills into supported coding agents.",
		Example:       "  skill-switcher source list\n  skill-switcher refresh\n  skill-switcher catalog list\n  skill-switcher sync --all",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          helpRunE,
	}

	rootCommand.SetOut(stdout)
	rootCommand.SetErr(stderr)

	rootCommand.AddGroup(
		&cobra.Group{ID: "workflow", Title: "Workflow Commands"},
		&cobra.Group{ID: "catalog", Title: "Catalog Commands"},
		&cobra.Group{ID: "source", Title: "Source Commands"},
	)

	rootCommand.AddCommand(
		newRefreshCommand(application),
		newSourceCommand(application),
		newCatalogCommand(application),
		newSyncCommand(application),
	)

	return rootCommand
}
