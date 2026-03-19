package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newRootCommand(stdout io.Writer, stderr io.Writer, application Application, openTUI func() error) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:           "skill-selector",
		Short:         "Manage shared skills across supported coding agents",
		Long:          "Skill Selector manages a shared skill catalog, refreshes upstream sources, and syncs selected skills into supported coding agents. Running `skill-selector` with no arguments opens the terminal UI.",
		Example:       "  skill-selector\n  skill-selector source list\n  skill-selector profile list\n  skill-selector refresh\n  skill-selector catalog list\n  skill-selector sync --all",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if openTUI == nil {
				return fmt.Errorf("tui launcher required")
			}

			return openTUI()
		},
	}

	rootCommand.SetOut(stdout)
	rootCommand.SetErr(stderr)

	rootCommand.AddGroup(
		&cobra.Group{ID: "workflow", Title: "Workflow Commands"},
		&cobra.Group{ID: "interface", Title: "Interface Commands"},
		&cobra.Group{ID: "catalog", Title: "Catalog Commands"},
		&cobra.Group{ID: "profile", Title: "Profile Commands"},
		&cobra.Group{ID: "source", Title: "Source Commands"},
	)

	rootCommand.AddCommand(
		newTUICommand(openTUI),
		newRefreshCommand(application),
		newSourceCommand(application),
		newProfileCommand(application),
		newCatalogCommand(application),
		newSyncCommand(application),
	)

	return rootCommand
}

func newTUICommand(openTUI func() error) *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Aliases: []string{"open"},
		Short:   "Open the terminal UI",
		Long:    "Open the Bubble Tea terminal UI for browsing sources, drafting a desired selection, and syncing it into supported agents.",
		Args:    cobra.NoArgs,
		GroupID: "interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			if openTUI == nil {
				return fmt.Errorf("tui launcher required")
			}

			return openTUI()
		},
	}
}
