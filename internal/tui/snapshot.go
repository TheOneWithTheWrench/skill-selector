package tui

import (
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/paths"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// Snapshot is the TUI read model loaded from the shared app layer.
type Snapshot struct {
	Runtime         paths.Runtime
	Sources         source.Sources
	Catalog         catalog.Catalog
	Profiles        profile.Profiles
	Manifests       []skillsync.Manifest
	SyncedSelection skill_identity.Identities
	Warnings        []string
}

func (s Snapshot) ActiveProfile() profile.Profile {
	return s.Profiles.Active()
}

func (s Snapshot) ActiveSelection() skill_identity.Identities {
	return s.ActiveProfile().Selected()
}

func newSnapshot(runtime paths.Runtime, configuredSources source.Sources, currentCatalog catalog.Catalog, profiles profile.Profiles, manifests []skillsync.Manifest) Snapshot {
	syncedSelection := projectSyncedSelection(manifests)
	warnings := projectWarnings(configuredSources, profiles.Active().Selected(), syncedSelection, manifests)

	return Snapshot{
		Runtime:         runtime,
		Sources:         configuredSources,
		Catalog:         currentCatalog,
		Profiles:        profiles,
		Manifests:       append([]skillsync.Manifest(nil), manifests...),
		SyncedSelection: syncedSelection,
		Warnings:        append([]string(nil), warnings...),
	}
}

func projectSyncedSelection(manifests []skillsync.Manifest) skill_identity.Identities {
	var syncedSelection skill_identity.Identities
	for _, manifest := range manifests {
		syncedSelection = append(syncedSelection, manifest.Identities()...)
	}

	return skill_identity.NewIdentities(syncedSelection...)
}

func projectWarnings(configuredSources source.Sources, activeSelection skill_identity.Identities, syncedSelection skill_identity.Identities, manifests []skillsync.Manifest) []string {
	var warnings []string
	if manifestsDiverge(manifests) {
		warnings = append(warnings, "Sync targets disagree on the current synced skills; the status pane uses the union.")
	}

	missingSourceIDs := missingSourceIDs(configuredSources, activeSelection)
	if len(missingSourceIDs) > 0 {
		warnings = append(warnings, fmt.Sprintf("The active profile selects skills from removed sources: %s.", summarizeIDs(missingSourceIDs)))
	}
	if !identitiesEqual(activeSelection, syncedSelection) {
		warnings = append(warnings, "The active profile selection differs from the current synced state. Press s to sync.")
	}

	return warnings
}

func manifestsDiverge(manifests []skillsync.Manifest) bool {
	if len(manifests) < 2 {
		return false
	}

	baseline := manifests[0].Identities()
	for _, manifest := range manifests[1:] {
		if !identitiesEqual(baseline, manifest.Identities()) {
			return true
		}
	}

	return false
}

func identitiesEqual(left skill_identity.Identities, right skill_identity.Identities) bool {
	left = skill_identity.NewIdentities(left...)
	right = skill_identity.NewIdentities(right...)
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index].Key() != right[index].Key() {
			return false
		}
	}

	return true
}

func missingSourceIDs(configuredSources source.Sources, identities skill_identity.Identities) []string {
	configuredSourceIDs := make(map[string]struct{}, len(configuredSources))
	for _, configuredSource := range configuredSources {
		configuredSourceIDs[configuredSource.ID()] = struct{}{}
	}

	seen := make(map[string]struct{})
	missing := make([]string, 0)
	for _, identity := range identities {
		if _, ok := configuredSourceIDs[identity.SourceID()]; ok {
			continue
		}
		if _, ok := seen[identity.SourceID()]; ok {
			continue
		}

		seen[identity.SourceID()] = struct{}{}
		missing = append(missing, identity.SourceID())
	}

	return missing
}

func summarizeIDs(values []string) string {
	if len(values) == 0 {
		return ""
	}

	if len(values) <= 3 {
		return strings.Join(values, ", ")
	}

	return strings.Join(values[:3], ", ") + fmt.Sprintf(" and %d more", len(values)-3)
}
