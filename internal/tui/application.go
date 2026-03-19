package tui

import (
	"context"

	"github.com/TheOneWithTheWrench/skill-selector/internal/app"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// Application is the core app surface used by the TUI.
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
	SaveActiveProfileSelection(skill_identity.Identities) (profile.Profiles, error)
	SyncSkillIdentities(skill_identity.Identities) (skillsync.Result, error)
	ListSyncManifests() ([]skillsync.Manifest, error)
}

// Workflow is the TUI-local command surface used by the Bubble Tea model.
type Workflow interface {
	AddSource(context.Context, string) (SourceActionResult, error)
	RemoveSource(context.Context, string) (SourceActionResult, error)
	Refresh(context.Context) (RefreshActionResult, error)
	CreateProfile(context.Context, string) (ProfilesActionResult, error)
	RenameProfile(context.Context, string, string) (ProfilesActionResult, error)
	RemoveProfile(context.Context, string) (ProfilesActionResult, error)
	SwitchProfile(context.Context, string) (ProfilesActionResult, error)
	Sync(context.Context, skill_identity.Identities) (SyncActionResult, error)
}

// SourceActionResult returns the reloaded TUI snapshot after a source mutation.
type SourceActionResult struct {
	Snapshot *Snapshot
	Source   source.Source
	Summary  string
}

// RefreshActionResult returns the reloaded TUI snapshot after a refresh.
type RefreshActionResult struct {
	Snapshot *Snapshot
	Summary  string
}

// ProfilesActionResult returns the reloaded TUI snapshot after one profile mutation.
type ProfilesActionResult struct {
	Snapshot *Snapshot
	Summary  string
}

// SyncActionResult returns the latest persisted sync state after one sync attempt.
type SyncActionResult struct {
	Snapshot *Snapshot
	Result   skillsync.Result
}
