package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/paths"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
)

// Service adapts the shared app use cases into the workflows expected by the TUI.
type Service struct {
	runtime     paths.Runtime
	application Application
}

// NewService binds the TUI workflows to one runtime and application surface.
func NewService(runtime paths.Runtime, application Application) (*Service, error) {
	if application == nil {
		return nil, fmt.Errorf("tui application required")
	}

	return &Service{
		runtime:     runtime,
		application: application,
	}, nil
}

// Load reads the current persisted state needed to start or reload the TUI.
func (s Service) Load(_ context.Context) (Snapshot, error) {
	configuredSources, err := s.application.ListSources()
	if err != nil {
		return Snapshot{}, err
	}

	currentCatalog, err := s.application.ListCatalog()
	if err != nil {
		return Snapshot{}, err
	}

	profiles, err := s.application.ListProfiles()
	if err != nil {
		return Snapshot{}, err
	}

	manifests, err := s.application.ListSyncManifests()
	if err != nil {
		return Snapshot{}, err
	}

	return newSnapshot(s.runtime, configuredSources, currentCatalog, profiles, manifests), nil
}

// AddSource adds one source, refreshes the catalog, and reloads the TUI snapshot.
func (s Service) AddSource(ctx context.Context, locator string) (SourceActionResult, error) {
	_, configuredSource, err := s.application.AddSource(locator)
	if err != nil {
		return SourceActionResult{}, err
	}

	refreshErr := s.refreshCatalog(ctx)
	snapshot, loadErr := s.loadSnapshot(ctx)

	return SourceActionResult{
		Snapshot: snapshot,
		Source:   configuredSource,
		Summary:  summarizeSourceAction("Added", configuredSource, snapshot),
	}, errors.Join(refreshErr, loadErr)
}

// RemoveSource removes one source, refreshes the catalog, and reloads the TUI snapshot.
func (s Service) RemoveSource(ctx context.Context, identifier string) (SourceActionResult, error) {
	_, removedSource, err := s.application.RemoveSource(identifier)
	if err != nil {
		return SourceActionResult{}, err
	}

	refreshErr := s.refreshCatalog(ctx)
	snapshot, loadErr := s.loadSnapshot(ctx)

	return SourceActionResult{
		Snapshot: snapshot,
		Source:   removedSource,
		Summary:  summarizeSourceAction("Removed", removedSource, snapshot),
	}, errors.Join(refreshErr, loadErr)
}

// Refresh refreshes every source mirror, rebuilds the catalog, and reloads the TUI snapshot.
func (s Service) Refresh(ctx context.Context) (RefreshActionResult, error) {
	refreshResult, refreshErr := s.application.RefreshCatalog(ctx)
	snapshot, loadErr := s.loadSnapshot(ctx)

	skillCount := len(refreshResult.Catalog.Skills())
	if snapshot != nil {
		skillCount = len(snapshot.Catalog.Skills())
	}

	return RefreshActionResult{
		Snapshot: snapshot,
		Summary: fmt.Sprintf(
			"Refreshed %d %s • indexed %d %s",
			len(refreshResult.Sources),
			pluralize(len(refreshResult.Sources), "source", "sources"),
			skillCount,
			pluralize(skillCount, "skill", "skills"),
		),
	}, errors.Join(refreshErr, loadErr)
}

// CreateProfile adds a new empty profile and reloads the TUI snapshot.
func (s Service) CreateProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	nextProfiles, err := s.application.CreateProfile(name)
	if err != nil {
		return ProfilesActionResult{}, err
	}

	snapshot, loadErr := s.loadSnapshot(ctx)
	return ProfilesActionResult{
		Snapshot: snapshot,
		Summary:  summarizeProfileAction("Created", nextProfiles, name),
	}, loadErr
}

// RenameProfile renames one profile and reloads the TUI snapshot.
func (s Service) RenameProfile(ctx context.Context, currentName string, newName string) (ProfilesActionResult, error) {
	nextProfiles, err := s.application.RenameProfile(currentName, newName)
	if err != nil {
		return ProfilesActionResult{}, err
	}

	snapshot, loadErr := s.loadSnapshot(ctx)
	return ProfilesActionResult{
		Snapshot: snapshot,
		Summary:  summarizeProfileAction("Renamed", nextProfiles, newName),
	}, loadErr
}

// RemoveProfile removes one inactive profile and reloads the TUI snapshot.
func (s Service) RemoveProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	_, err := s.application.RemoveProfile(name)
	if err != nil {
		return ProfilesActionResult{}, err
	}

	snapshot, loadErr := s.loadSnapshot(ctx)
	return ProfilesActionResult{
		Snapshot: snapshot,
		Summary:  fmt.Sprintf("Removed profile %s", normalizeProfileName(name)),
	}, loadErr
}

// SwitchProfile changes the active profile and reloads the TUI snapshot without syncing it automatically.
func (s Service) SwitchProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	nextProfiles, err := s.application.SwitchProfile(name)
	if err != nil {
		return ProfilesActionResult{}, err
	}

	snapshot, loadErr := s.loadSnapshot(ctx)
	return ProfilesActionResult{
		Snapshot: snapshot,
		Summary:  fmt.Sprintf("Switched active profile to %s", nextProfiles.Active().Name()),
	}, loadErr
}

// Sync reconciles the desired selection and reloads the persisted sync state.

func (s Service) Sync(ctx context.Context, desired skill_identity.Identities) (SyncActionResult, error) {
	if _, err := s.application.SaveActiveProfileSelection(desired); err != nil {
		return SyncActionResult{}, err
	}

	result, syncErr := s.application.SyncSkillIdentities(desired)
	snapshot, loadErr := s.loadSnapshot(ctx)

	return SyncActionResult{
		Snapshot: snapshot,
		Result:   result,
	}, errors.Join(syncErr, loadErr)
}

func (s Service) loadSnapshot(ctx context.Context) (*Snapshot, error) {
	snapshot, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (s Service) refreshCatalog(ctx context.Context) error {
	_, err := s.application.RefreshCatalog(ctx)
	return err
}

func summarizeSourceAction(action string, configuredSource source.Source, snapshot *Snapshot) string {
	summary := fmt.Sprintf("%s %s", action, configuredSource.Locator())
	if snapshot == nil {
		return summary
	}

	skillCount := len(snapshot.Catalog.Skills())
	return fmt.Sprintf(
		"%s • indexed %d %s",
		summary,
		skillCount,
		pluralize(skillCount, "skill", "skills"),
	)
}

func summarizeProfileAction(action string, profiles profile.Profiles, name string) string {
	item, ok := profiles.Find(name)
	if !ok {
		return fmt.Sprintf("%s profile %s", action, normalizeProfileName(name))
	}

	return fmt.Sprintf("%s profile %s", action, item.Name())
}

func normalizeProfileName(name string) string {
	normalizedName := strings.TrimSpace(name)
	if strings.EqualFold(normalizedName, profile.DefaultName) {
		return profile.DefaultName
	}

	return normalizedName
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}
