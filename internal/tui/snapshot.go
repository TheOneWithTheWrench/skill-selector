package tui

import (
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
)

// Snapshot is the TUI read model loaded from the shared app layer.
type Snapshot struct {
	Runtime         paths.Runtime
	Sources         source.Sources
	Catalog         catalog.Catalog
	Manifests       []skillsync.Manifest
	ActiveSelection skillidentity.Identities
	Warnings        []string
}

func newSnapshot(runtime paths.Runtime, configuredSources source.Sources, currentCatalog catalog.Catalog, manifests []skillsync.Manifest) Snapshot {
	activeSelection, warnings := projectActiveSelection(configuredSources, manifests)

	return Snapshot{
		Runtime:         runtime,
		Sources:         configuredSources,
		Catalog:         currentCatalog,
		Manifests:       append([]skillsync.Manifest(nil), manifests...),
		ActiveSelection: activeSelection,
		Warnings:        append([]string(nil), warnings...),
	}
}

func projectActiveSelection(configuredSources source.Sources, manifests []skillsync.Manifest) (skillidentity.Identities, []string) {
	activeSelection := make(skillidentity.Identities, 0)
	for _, manifest := range manifests {
		activeSelection = append(activeSelection, manifest.Identities()...)
	}

	activeSelection = skillidentity.NewIdentities(activeSelection...)

	var warnings []string
	if manifestsDiverge(manifests) {
		warnings = append(warnings, "Sync targets disagree on the active skills; the TUI uses the union.")
	}

	missingSourceIDs := missingSourceIDs(configuredSources, activeSelection)
	if len(missingSourceIDs) > 0 {
		warnings = append(warnings, fmt.Sprintf("Some synced skills belong to removed sources: %s. Sync to clear them.", summarizeIDs(missingSourceIDs)))
	}

	return activeSelection, warnings
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

func identitiesEqual(left skillidentity.Identities, right skillidentity.Identities) bool {
	left = skillidentity.NewIdentities(left...)
	right = skillidentity.NewIdentities(right...)
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

func missingSourceIDs(configuredSources source.Sources, identities skillidentity.Identities) []string {
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
