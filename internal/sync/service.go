package sync

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

type targetGroup struct {
	Target   Target
	Adapters []string
	Manifest Manifest
}

// Run reconciles desired skill identities across every target and returns updated manifests.
func Run(desired skill_identity.Identities, targets []Target, manifests []Manifest, resolve Resolver) (Result, error) {
	desired = skill_identity.NewIdentities(desired...)
	targetGroups := groupTargets(targets, manifests)

	var (
		result    = Result{DesiredCount: len(desired)}
		allErrors []error
	)

	for _, targetGroup := range targetGroups {
		targetResult, updatedManifest, err := SyncTarget(desired, targetGroup.Target, targetGroup.Manifest, resolve)
		targetResult.Adapters = append([]string(nil), targetGroup.Adapters...)
		targetResult.RootPath = targetGroup.Target.RootPath()
		if len(targetGroup.Adapters) > 1 {
			targetResult.Adapter = strings.Join(targetGroup.Adapters, ",")
		}
		result.Targets = append(result.Targets, targetResult)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}

		for _, adapter := range targetGroup.Adapters {
			result.Manifests = append(result.Manifests, updatedManifest.withAdapter(adapter))
		}
	}

	slices.SortFunc(result.Targets, func(left TargetResult, right TargetResult) int {
		return strings.Compare(left.Adapter, right.Adapter)
	})
	slices.SortFunc(result.Manifests, func(left Manifest, right Manifest) int {
		return strings.Compare(left.Adapter(), right.Adapter())
	})

	return result, errors.Join(allErrors...)
}

// SyncTarget reconciles one target against desired identities and the target's last manifest.
func SyncTarget(desired skill_identity.Identities, target Target, manifest Manifest, resolve Resolver) (TargetResult, Manifest, error) {
	result := TargetResult{Adapter: target.Adapter()}

	if target.Adapter() == "" {
		err := fmt.Errorf("target adapter required")
		result.Error = err.Error()
		return result, manifest, err
	}
	if target.RootPath() == "" {
		err := fmt.Errorf("target root path required for %q", target.Adapter())
		result.Error = err.Error()
		return result, manifest, err
	}
	if resolve == nil {
		err := fmt.Errorf("target resolver required for %q", target.Adapter())
		result.Error = err.Error()
		return result, manifest, err
	}

	manifest = manifest.withAdapter(target.Adapter()).withRootPath(target.RootPath())
	desired = skill_identity.NewIdentities(desired...)
	desiredIndex := make(map[string]skill_identity.Identity, len(desired))
	for _, identity := range desired {
		desiredIndex[identity.Key()] = identity
	}

	var linkedIdentities skill_identity.Identities
	for _, identity := range desired {
		sourcePath, err := resolve(identity)
		if errors.Is(err, os.ErrNotExist) {
			removed, removeErr := removeOwnedLink(target.LinkPath(identity))
			if removeErr != nil {
				wrapped := fmt.Errorf("remove stale link for missing %q on %q: %w", identity.Key(), target.Adapter(), removeErr)
				result.Error = wrapped.Error()
				return result, manifest, wrapped
			}
			if removed {
				result.Removed++
			}
			result.Skipped++
			continue
		}
		if err != nil {
			wrapped := fmt.Errorf("resolve source for %q on %q: %w", identity.Key(), target.Adapter(), err)
			result.Error = wrapped.Error()
			return result, manifest, wrapped
		}

		info, err := os.Stat(sourcePath)
		if errors.Is(err, os.ErrNotExist) {
			removed, removeErr := removeOwnedLink(target.LinkPath(identity))
			if removeErr != nil {
				wrapped := fmt.Errorf("remove stale link for missing %q on %q: %w", identity.Key(), target.Adapter(), removeErr)
				result.Error = wrapped.Error()
				return result, manifest, wrapped
			}
			if removed {
				result.Removed++
			}
			result.Skipped++
			continue
		}
		if err != nil {
			wrapped := fmt.Errorf("stat source path %q: %w", sourcePath, err)
			result.Error = wrapped.Error()
			return result, manifest, wrapped
		}
		if !info.IsDir() {
			wrapped := fmt.Errorf("skill source is not a directory: %q", sourcePath)
			result.Error = wrapped.Error()
			return result, manifest, wrapped
		}

		changed, err := ensureSymlink(sourcePath, target.LinkPath(identity))
		if err != nil {
			wrapped := fmt.Errorf("sync %q on %q: %w", identity.Key(), target.Adapter(), err)
			result.Error = wrapped.Error()
			return result, manifest, wrapped
		}
		if changed {
			result.Linked++
		} else {
			result.Unchanged++
		}

		linkedIdentities = append(linkedIdentities, identity)
	}

	for _, identity := range manifest.Identities() {
		if _, ok := desiredIndex[identity.Key()]; ok {
			continue
		}

		removed, err := removeOwnedLink(target.LinkPath(identity))
		if err != nil {
			wrapped := fmt.Errorf("remove stale link for %q on %q: %w", identity.Key(), target.Adapter(), err)
			result.Error = wrapped.Error()
			return result, manifest, wrapped
		}
		if removed {
			result.Removed++
		}
	}

	manifest = manifest.withIdentities(linkedIdentities)

	return result, manifest, nil
}

func groupTargets(targets []Target, manifests []Manifest) []targetGroup {
	manifestIndex := make(map[string]Manifest, len(manifests))
	for _, manifest := range manifests {
		manifestIndex[manifest.Adapter()] = manifest
	}

	sortedTargets := append([]Target(nil), targets...)
	slices.SortFunc(sortedTargets, func(left Target, right Target) int {
		if left.RootPath() == right.RootPath() {
			return strings.Compare(left.Adapter(), right.Adapter())
		}

		return strings.Compare(left.RootPath(), right.RootPath())
	})

	groupIndex := make(map[string]int, len(sortedTargets))
	groups := make([]targetGroup, 0, len(sortedTargets))

	for _, target := range sortedTargets {
		rootKey := target.RootPath()
		if rootKey == "" {
			rootKey = target.Adapter()
		}

		manifest, ok := manifestIndex[target.Adapter()]
		if !ok {
			manifest, _ = NewManifest(target.Adapter(), target.RootPath())
		}
		manifest = manifest.withRootPath(target.RootPath())

		if index, ok := groupIndex[rootKey]; ok {
			groups[index].Adapters = append(groups[index].Adapters, target.Adapter())
			groups[index].Manifest = groups[index].Manifest.withIdentities(append(groups[index].Manifest.Identities(), manifest.Identities()...))
			continue
		}

		groups = append(groups, targetGroup{
			Target:   target,
			Adapters: []string{target.Adapter()},
			Manifest: manifest,
		})
		groupIndex[rootKey] = len(groups) - 1
	}

	for index := range groups {
		slices.Sort(groups[index].Adapters)
	}

	return groups
}

func (m Manifest) withAdapter(adapter string) Manifest {
	m.adapter = strings.TrimSpace(adapter)
	return m
}
