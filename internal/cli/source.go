package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newSourceCommand(application Application) *cobra.Command {
	sourceCommand := &cobra.Command{
		Use:     "source",
		Short:   "Manage configured skill sources",
		Long:    "Manage the upstream skill sources that are cloned, scanned, and later synced into supported agents.",
		GroupID: "source",
		RunE:    helpRunE,
	}

	sourceCommand.AddCommand(
		newSourceListCommand(application),
		newSourceAddCommand(application),
		newSourceRemoveCommand(application),
	)

	return sourceCommand
}

func newSourceListCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured skill sources",
		Long:    "List every configured source with its stable source ID and user-facing locator.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configuredSources, err := application.ListSources()
			if err != nil {
				return err
			}

			if len(configuredSources) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No sources configured.")
				return err
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, configuredSource := range configuredSources {
				if _, err := fmt.Fprintf(writer, "%s\t%s\n", configuredSource.ID(), configuredSource.Locator()); err != nil {
					return err
				}
			}

			return writer.Flush()
		},
	}
}

func newSourceAddCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "add <locator>",
		Short: "Add a skill source",
		Long:  "Add a supported source locator, such as a GitHub tree URL, to the configured source list.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, configuredSource, err := application.AddSource(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added %s\n", configuredSource.Locator())
			return err
		},
	}
}

func newSourceRemoveCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <locator-or-id>",
		Short: "Remove a configured skill source",
		Long:  "Remove a configured source by its stable source ID or by the exact locator that was originally added.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, removedSource, err := application.RemoveSource(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", removedSource.Locator())
			return err
		},
	}
}
