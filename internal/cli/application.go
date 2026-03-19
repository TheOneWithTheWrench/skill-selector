package cli

import (
	"context"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
)

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
