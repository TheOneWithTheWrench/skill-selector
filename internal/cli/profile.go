package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/spf13/cobra"
)

func newProfileCommand(application Application) *cobra.Command {
	profileCommand := &cobra.Command{
		Use:     "profile",
		Short:   "Manage saved skill profiles",
		Long:    "Manage the named saved selections that the TUI edits and syncs explicitly.",
		GroupID: "profile",
		RunE:    helpRunE,
	}

	profileCommand.AddCommand(
		newProfileListCommand(application),
		newProfileCreateCommand(application),
		newProfileRenameCommand(application),
		newProfileRemoveCommand(application),
		newProfileSwitchCommand(application),
	)

	return profileCommand
}

func newProfileListCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved profiles",
		Long:  "List every saved profile together with its active marker and selected skill count.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := application.ListProfiles()
			if err != nil {
				return err
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, item := range profiles.All() {
				marker := " "
				if item.Name() == profiles.ActiveName() {
					marker = "*"
				}

				if _, err := fmt.Fprintf(writer, "%s\t%s\t%d %s\n", marker, item.Name(), item.SelectedCount(), pluralize(item.SelectedCount(), "skill", "skills")); err != nil {
					return err
				}
			}

			return writer.Flush()
		},
	}
}

func newProfileCreateCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a saved profile",
		Long:  "Create a new empty saved profile without switching to it.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := application.CreateProfile(args[0])
			if err != nil {
				return err
			}

			createdProfile, ok := profiles.Find(args[0])
			if !ok {
				createdProfile = profile.Default()
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created profile %s\n", createdProfile.Name())
			return err
		},
	}
}

func newProfileRenameCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <current-name> <new-name>",
		Short: "Rename a saved profile",
		Long:  "Rename one saved profile while keeping its selection and active status intact.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := application.RenameProfile(args[0], args[1])
			if err != nil {
				return err
			}

			renamedProfile, ok := profiles.Find(args[1])
			if !ok {
				renamedProfile = profile.Default()
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Renamed profile to %s\n", renamedProfile.Name())
			return err
		},
	}
}

func newProfileRemoveCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a saved profile",
		Long:  "Remove one inactive saved profile and keep the remaining saved selections unchanged.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := application.RemoveProfile(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed profile %s\n", strings.TrimSpace(args[0]))
			return err
		},
	}
}

func newProfileSwitchCommand(application Application) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch the active saved profile",
		Long:  "Switch the active saved profile without syncing it automatically.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := application.SwitchProfile(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched active profile to %s\n", profiles.Active().Name())
			return err
		},
	}
}
