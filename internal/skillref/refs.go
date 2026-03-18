package skillref

import (
	"fmt"
	"slices"
	"strings"
)

// Refs is the normalized collection of lightweight skill references.
// Nil and zero-length collections are both valid empty values.
type Refs []Ref

// NewRefs sorts refs into stable order and removes duplicate keys.
func NewRefs(items ...Ref) Refs {
	var normalizedItems Refs
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if _, ok := seen[item.Key()]; ok {
			continue
		}

		seen[item.Key()] = struct{}{}
		normalizedItems = append(normalizedItems, item)
	}

	slices.SortFunc(normalizedItems, func(left Ref, right Ref) int {
		if left.SourceID() == right.SourceID() {
			return strings.Compare(left.RelativePath(), right.RelativePath())
		}

		return strings.Compare(left.SourceID(), right.SourceID())
	})

	return normalizedItems
}

// Add returns a new normalized collection with one extra ref.
func (r Refs) Add(candidate Ref) (Refs, error) {
	for _, existing := range r {
		if existing.Key() == candidate.Key() {
			return nil, fmt.Errorf("skill ref already exists: %s", candidate.Key())
		}
	}

	nextRefs := append(append(Refs(nil), r...), candidate)
	return NewRefs(nextRefs...), nil
}

// Remove drops one ref by stable key and returns the next collection.
func (r Refs) Remove(candidate Ref) (Refs, error) {
	index := -1
	for currentIndex, current := range r {
		if current.Key() == candidate.Key() {
			index = currentIndex
			break
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("skill ref not found: %s", candidate.Key())
	}

	nextRefs := append(append(Refs(nil), r[:index]...), r[index+1:]...)
	return NewRefs(nextRefs...), nil
}
