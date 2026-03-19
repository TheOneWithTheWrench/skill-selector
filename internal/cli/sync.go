package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCommand(application Application) *cobra.Command {
	var all bool

	syncCommand := &cobra.Command{
		Use:     "sync [skill-identity...]",
		Short:   "Sync selected skills into supported agents",
		Long:    "Sync one or more skill identities, or the entire catalog, into every configured agent target while preserving per-target manifests.",
		Example: "  skill-switcher sync source-id:reviewer\n  skill-switcher sync --all\n  skill-switcher sync clear\n  skill-switcher sync status",
		Args:    cobra.ArbitraryArgs,
		GroupID: "workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			var identities skillidentity.Identities

			if all {
				if len(args) > 0 {
					return fmt.Errorf("sync --all does not accept explicit skill identities")
				}

				currentCatalog, err := application.ListCatalog()
				if err != nil {
					return err
				}

				for _, discoveredSkill := range currentCatalog.Skills() {
					identities = append(identities, discoveredSkill.Identity())
				}

				if len(identities) == 0 {
					return fmt.Errorf("catalog is empty; run `skill-switcher refresh` first")
				}

				identities = skillidentity.NewIdentities(identities...)
			} else {
				if len(args) == 0 {
					return fmt.Errorf("missing skill identity")
				}

				parsedIdentities := make(skillidentity.Identities, 0, len(args))
				for _, rawIdentity := range args {
					identity, err := skillidentity.Parse(rawIdentity)
					if err != nil {
						return err
					}

					parsedIdentities = append(parsedIdentities, identity)
				}

				identities = skillidentity.NewIdentities(parsedIdentities...)
			}

			result, err := application.SyncSkillIdentities(identities)
			if writeErr := writeSyncResult(cmd.OutOrStdout(), result); writeErr != nil {
				return writeErr
			}

			return err
		},
	}

	syncCommand.Flags().BoolVar(&all, "all", false, "sync every skill currently present in the catalog")

	syncCommand.AddCommand(
		newSyncStatusCommand(application),
		newSyncClearCommand(application),
	)

	return syncCommand
}

func newSyncStatusCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show persisted sync manifests",
		Long:  "Show the persisted per-target sync manifests that describe which skill identities are currently owned by each target.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifests, err := application.ListSyncManifests()
			if err != nil {
				return err
			}

			if len(manifests) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No sync manifests found.")
				return err
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, manifest := range manifests {
				if _, err := fmt.Fprintf(writer, "%s\t%s\t%d %s\n", manifest.Adapter(), manifest.RootPath(), len(manifest.Identities()), pluralize(len(manifest.Identities()), "skill", "skills")); err != nil {
					return err
				}
			}

			return writer.Flush()
		},
	}
}

func newSyncClearCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove all synced skills from every target",
		Long:  "Clear every synced skill from every configured target while keeping manifests in sync with the empty desired state.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := application.SyncSkillIdentities(nil)
			if writeErr := writeSyncResult(cmd.OutOrStdout(), result); writeErr != nil {
				return writeErr
			}

			return err
		},
	}
}

func writeSyncResult(stdout io.Writer, result skillsync.Result) error {
	if _, err := fmt.Fprintln(stdout, result.Summary()); err != nil {
		return err
	}

	writer := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	for _, target := range result.Targets {
		line := fmt.Sprintf("linked=%d\tremoved=%d\tskipped=%d\tunchanged=%d", target.Linked, target.Removed, target.Skipped, target.Unchanged)
		if target.Error != "" {
			line += "\terror=" + target.Error
		}

		if _, err := fmt.Fprintf(writer, "%s\t%s\n", target.DisplayName(), line); err != nil {
			return err
		}
	}

	return writer.Flush()
}
