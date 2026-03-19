package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

type syncLocationState struct {
	RootPath   string
	Adapters   []string
	Identities skill_identity.Identities
}

func (m Model) syncLocationStates() []syncLocationState {
	type groupedState struct {
		RootPath   string
		Adapters   []string
		Identities skill_identity.Identities
	}

	groupedStates := make(map[string]*groupedState, len(m.snapshot.Manifests))
	for _, manifest := range m.snapshot.Manifests {
		key := manifest.RootPath()
		if key == "" {
			key = manifest.Adapter()
		}

		group, ok := groupedStates[key]
		if !ok {
			group = &groupedState{RootPath: manifest.RootPath()}
			groupedStates[key] = group
		}

		group.Adapters = append(group.Adapters, manifest.Adapter())
		group.Identities = append(group.Identities, manifest.Identities()...)
	}

	result := make([]syncLocationState, 0, len(groupedStates))
	for _, groupedState := range groupedStates {
		adapters := append([]string(nil), groupedState.Adapters...)
		slices.Sort(adapters)

		result = append(result, syncLocationState{
			RootPath:   groupedState.RootPath,
			Adapters:   adapters,
			Identities: skill_identity.NewIdentities(groupedState.Identities...),
		})
	}

	slices.SortFunc(result, func(left syncLocationState, right syncLocationState) int {
		leftKey := left.RootPath + strings.Join(left.Adapters, ",")
		rightKey := right.RootPath + strings.Join(right.Adapters, ",")
		return strings.Compare(leftKey, rightKey)
	})

	return result
}

func renderSyncLocationDetail(syncLocation syncLocationState) string {
	rootPath := syncLocation.RootPath
	if rootPath == "" {
		rootPath = "auto-detected at runtime"
	}

	lines := []string{
		"Location: " + rootPath,
		"Adapters: " + strings.Join(syncLocation.Adapters, ", "),
		fmt.Sprintf("Synced skills: %d", len(syncLocation.Identities)),
	}

	if len(syncLocation.Identities) == 0 {
		lines = append(lines, "", "Run sync once there is a desired selection.")
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")
	for _, identity := range syncLocation.Identities {
		lines = append(lines, "- "+identity.SourceID()+" • "+displayRelativePath(identity.RelativePath()))
	}

	return strings.Join(lines, "\n")
}

func manifestsFromResult(targets []skillsync.Manifest) []skillsync.Manifest {
	return append([]skillsync.Manifest(nil), targets...)
}
