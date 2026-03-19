package skill_identity

import (
	"fmt"
	"slices"
	"strings"
)

// Identities is the normalized collection of lightweight skill identities.
// Nil and zero-length collections are both valid empty values.
type Identities []Identity

// NewIdentities sorts identities into stable order and removes duplicate keys.
func NewIdentities(items ...Identity) Identities {
	var normalizedItems Identities
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if _, ok := seen[item.Key()]; ok {
			continue
		}

		seen[item.Key()] = struct{}{}
		normalizedItems = append(normalizedItems, item)
	}

	slices.SortFunc(normalizedItems, func(left Identity, right Identity) int {
		if left.SourceID() == right.SourceID() {
			return strings.Compare(left.RelativePath(), right.RelativePath())
		}

		return strings.Compare(left.SourceID(), right.SourceID())
	})

	return normalizedItems
}

// Add returns a new normalized collection with one extra identity.
func (i Identities) Add(candidate Identity) (Identities, error) {
	for _, existing := range i {
		if existing.Key() == candidate.Key() {
			return nil, fmt.Errorf("skill identity already exists: %s", candidate.Key())
		}
	}

	nextIdentities := append(append(Identities(nil), i...), candidate)
	return NewIdentities(nextIdentities...), nil
}

// Remove drops one identity by stable key and returns the next collection.
func (i Identities) Remove(candidate Identity) (Identities, error) {
	index := -1
	for currentIndex, current := range i {
		if current.Key() == candidate.Key() {
			index = currentIndex
			break
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("skill identity not found: %s", candidate.Key())
	}

	nextIdentities := append(append(Identities(nil), i[:index]...), i[index+1:]...)
	return NewIdentities(nextIdentities...), nil
}
