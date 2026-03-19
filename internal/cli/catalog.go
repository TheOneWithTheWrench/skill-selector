package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newCatalogCommand(application Application) *cobra.Command {
	catalogCommand := &cobra.Command{
		Use:     "catalog",
		Short:   "Inspect the discovered skill catalog",
		Long:    "Inspect the local catalog snapshot that was built from refreshed source mirrors.",
		GroupID: "catalog",
		RunE:    helpRunE,
	}

	catalogCommand.AddCommand(newCatalogListCommand(application))

	return catalogCommand
}

func newCatalogListCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered skills in the catalog",
		Long:  "List discovered skills by stable skill identity together with their human-readable catalog metadata.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentCatalog, err := application.ListCatalog()
			if err != nil {
				return err
			}

			discoveredSkills := currentCatalog.Skills()
			if len(discoveredSkills) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "Catalog is empty. Run `skill-switcher refresh` first.")
				return err
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, discoveredSkill := range discoveredSkills {
				if _, err := fmt.Fprintf(writer, "%s\t%s\n", discoveredSkill.Identity().Key(), discoveredSkill.Name()); err != nil {
					return err
				}
				if description := strings.TrimSpace(discoveredSkill.Description()); description != "" {
					if _, err := fmt.Fprintf(writer, "\t%s\n", description); err != nil {
						return err
					}
				}
			}

			return writer.Flush()
		},
	}
}
