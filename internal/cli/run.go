package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
)

const Usage = `usage: skill-switcher <command>

commands:
  source list
  source add <locator>
  source remove <locator|source-id>
  refresh
  pull
  catalog list
  sync status
  sync clear
  sync --all
  sync <skill-identity>...`

// Application is the core app surface used by the CLI.
type Application interface {
	ListSources() (source.Sources, error)
	AddSource(string) (source.Sources, source.Source, error)
	RemoveSource(string) (source.Sources, source.Source, error)
	RefreshCatalog(context.Context) (app.RefreshCatalogResult, error)
	ListCatalog() (catalog.Catalog, error)
	SyncSkillIdentities(skillidentity.Identities) (skillsync.Result, error)
	ListSyncManifests() ([]skillsync.Manifest, error)
}

// Run parses CLI arguments, calls the shared core, and writes text output.
func Run(args []string, stdout io.Writer, stderr io.Writer, application Application) error {
	if application == nil {
		return fmt.Errorf("cli application required")
	}

	if len(args) < 2 {
		_, err := fmt.Fprintln(stdout, Usage)
		return err
	}

	switch args[1] {
	case "-h", "--help", "help":
		_, err := fmt.Fprintln(stdout, Usage)
		return err
	case "refresh", "pull":
		return runRefresh(stdout, application)
	case "source":
		return runSource(args[2:], stdout, application)
	case "catalog":
		return runCatalog(args[2:], stdout, application)
	case "sync":
		return runSync(args[2:], stdout, stderr, application)
	default:
		if _, err := fmt.Fprintln(stdout, Usage); err != nil {
			return err
		}

		return fmt.Errorf("unknown command %q", args[1])
	}
}

func runSource(args []string, stdout io.Writer, application Application) error {
	if len(args) == 0 {
		return fmt.Errorf("missing source subcommand")
	}

	switch args[0] {
	case "list":
		configuredSources, err := application.ListSources()
		if err != nil {
			return err
		}

		if len(configuredSources) == 0 {
			_, err := fmt.Fprintln(stdout, "No sources configured.")
			return err
		}

		writer := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		for _, configuredSource := range configuredSources {
			if _, err := fmt.Fprintf(writer, "%s\t%s\n", configuredSource.ID(), configuredSource.Locator()); err != nil {
				return err
			}
		}

		return writer.Flush()
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("missing source locator")
		}

		_, configuredSource, err := application.AddSource(args[1])
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(stdout, "Added %s\n", configuredSource.Locator())
		return err
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("missing source identifier")
		}

		_, removedSource, err := application.RemoveSource(args[1])
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(stdout, "Removed %s\n", removedSource.Locator())
		return err
	default:
		return fmt.Errorf("unknown source subcommand %q", args[0])
	}
}

func runRefresh(stdout io.Writer, application Application) error {
	result, err := application.RefreshCatalog(context.Background())

	writer := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
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
	if _, writeErr := fmt.Fprintf(stdout, "Catalog contains %d %s.\n", len(result.Catalog.Skills()), pluralize(len(result.Catalog.Skills()), "skill", "skills")); writeErr != nil {
		return writeErr
	}

	return err
}

func runCatalog(args []string, stdout io.Writer, application Application) error {
	if len(args) == 0 {
		return fmt.Errorf("missing catalog subcommand")
	}

	if args[0] != "list" {
		return fmt.Errorf("unknown catalog subcommand %q", args[0])
	}

	currentCatalog, err := application.ListCatalog()
	if err != nil {
		return err
	}

	discoveredSkills := currentCatalog.Skills()
	if len(discoveredSkills) == 0 {
		_, err := fmt.Fprintln(stdout, "Catalog is empty. Run `skill-switcher refresh` first.")
		return err
	}

	writer := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
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
}

func runSync(args []string, stdout io.Writer, stderr io.Writer, application Application) error {
	if len(args) == 0 {
		return fmt.Errorf("missing sync command or skill identity")
	}

	switch args[0] {
	case "status":
		return runSyncStatus(stdout, application)
	case "clear":
		result, err := application.SyncSkillIdentities(nil)
		if writeErr := writeSyncResult(stdout, result); writeErr != nil {
			return writeErr
		}
		return err
	}

	flagSet := flag.NewFlagSet("sync", flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	all := flagSet.Bool("all", false, "sync every skill in the catalog")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	var identities skillidentity.Identities
	if *all {
		if flagSet.NArg() > 0 {
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
		if flagSet.NArg() == 0 {
			return fmt.Errorf("missing skill identity")
		}

		parsedIdentities := make(skillidentity.Identities, 0, flagSet.NArg())
		for _, rawIdentity := range flagSet.Args() {
			identity, err := skillidentity.Parse(rawIdentity)
			if err != nil {
				return err
			}

			parsedIdentities = append(parsedIdentities, identity)
		}

		identities = skillidentity.NewIdentities(parsedIdentities...)
	}

	result, err := application.SyncSkillIdentities(identities)
	if writeErr := writeSyncResult(stdout, result); writeErr != nil {
		return writeErr
	}

	return err
}

func runSyncStatus(stdout io.Writer, application Application) error {
	manifests, err := application.ListSyncManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		_, err := fmt.Fprintln(stdout, "No sync manifests found.")
		return err
	}

	writer := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	for _, manifest := range manifests {
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%d %s\n", manifest.Adapter(), manifest.RootPath(), len(manifest.Identities()), pluralize(len(manifest.Identities()), "skill", "skills")); err != nil {
			return err
		}
	}

	return writer.Flush()
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

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}
