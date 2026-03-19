package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newRefreshCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:     "refresh",
		Aliases: []string{"pull"},
		Short:   "Refresh sources and rebuild the catalog",
		Long:    "Refresh every configured source mirror, then rebuild the local catalog from the refreshed source trees.",
		Args:    cobra.NoArgs,
		GroupID: "workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := application.RefreshCatalog(context.Background())

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			if len(result.Sources) == 0 {
				if _, writeErr := fmt.Fprintln(writer, "No sources refreshed."); writeErr != nil {
					return writeErr
				}
			} else {
				for _, refreshedSource := range result.Sources {
					if _, writeErr := fmt.Fprintf(writer, "%s\t%s\t%s\n", refreshedSource.Action, refreshedSource.Mirror.ID(), refreshedSource.Mirror.Source.Locator()); writeErr != nil {
						return writeErr
					}
				}
			}
			if writeErr := writer.Flush(); writeErr != nil {
				return writeErr
			}

			if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Catalog contains %d %s.\n", len(result.Catalog.Skills()), pluralize(len(result.Catalog.Skills()), "skill", "skills")); writeErr != nil {
				return writeErr
			}

			return err
		},
	}
}
