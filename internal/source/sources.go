package source

import (
	"fmt"
	"slices"
	"strings"
)

// Sources is the normalized collection of configured sources.
// Nil and zero-length collections are both valid empty values.
type Sources []Source

// NewSources creates a collection sorted by URL and deduplicated by source ID.
func NewSources(items ...Source) Sources {
	return normalizeSources(items)
}

// Add returns a new collection with the source included, rejecting duplicate source IDs.
func (s Sources) Add(candidate Source) (Sources, error) {
	for _, existingSource := range s {
		if existingSource.ID() == candidate.ID() {
			return nil, fmt.Errorf("source already exists: %s", candidate.Locator())
		}
	}

	nextSources := append(s.clone(), candidate)

	return normalizeSources(nextSources), nil
}

// Remove matches a source by exact URL or stable source ID and returns the next collection.
func (s Sources) Remove(identifier string) (Sources, Source, error) {
	normalizedIdentifier := strings.TrimSpace(identifier)
	if normalizedIdentifier == "" {
		return nil, Source{}, fmt.Errorf("source identifier required")
	}

	index := -1
	removedSource := Source{}

	for candidateIndex, candidateSource := range s {
		if candidateSource.Locator() == normalizedIdentifier || candidateSource.ID() == normalizedIdentifier {
			index = candidateIndex
			removedSource = candidateSource
			break
		}
	}

	if index == -1 {
		return nil, Source{}, fmt.Errorf("source not found: %s", normalizedIdentifier)
	}

	nextSources := s.clone()
	nextSources = append(nextSources[:index], nextSources[index+1:]...)

	return normalizeSources(nextSources), removedSource, nil
}

func (s Sources) clone() Sources {
	return append(Sources(nil), s...)
}

func normalizeSources(items Sources) Sources {
	var normalizedItems Sources
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if _, ok := seen[item.ID()]; ok {
			continue
		}

		seen[item.ID()] = struct{}{}
		normalizedItems = append(normalizedItems, item)
	}

	slices.SortFunc(normalizedItems, func(left Source, right Source) int {
		return strings.Compare(left.Locator(), right.Locator())
	})

	return normalizedItems
}
