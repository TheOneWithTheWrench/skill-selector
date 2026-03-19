package cli

import (
	"context"

	"github.com/TheOneWithTheWrench/skill-selector/internal/app"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// Application is the core app surface used by the CLI.
type Application interface {
	ListSources() (source.Sources, error)
	AddSource(string) (source.Sources, source.Source, error)
	RemoveSource(string) (source.Sources, source.Source, error)
	RefreshCatalog(context.Context) (app.RefreshCatalogResult, error)
	ListCatalog() (catalog.Catalog, error)
	ListProfiles() (profile.Profiles, error)
	CreateProfile(string) (profile.Profiles, error)
	RenameProfile(string, string) (profile.Profiles, error)
	RemoveProfile(string) (profile.Profiles, error)
	SwitchProfile(string) (profile.Profiles, error)
	ActivateProfile(string) (app.ActivateProfileResult, error)
	SyncSkillIdentities(skill_identity.Identities) (skillsync.Result, error)
	ListSyncManifests() ([]skillsync.Manifest, error)
}
