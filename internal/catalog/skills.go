package catalog

import (
	"slices"
	"strings"
)

// Skills is the normalized collection of discovered skills.
// Nil and zero-length collections are both valid empty values.
type Skills []Skill

// NewSkills sorts skills into stable order and removes duplicate skill IDs.
func NewSkills(items ...Skill) Skills {
	var normalizedItems Skills
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if _, ok := seen[item.ID()]; ok {
			continue
		}

		seen[item.ID()] = struct{}{}
		normalizedItems = append(normalizedItems, item)
	}

	slices.SortFunc(normalizedItems, func(left Skill, right Skill) int {
		if left.SourceID() == right.SourceID() {
			if left.Name() == right.Name() {
				return strings.Compare(left.RelativePath(), right.RelativePath())
			}

			return strings.Compare(left.Name(), right.Name())
		}

		return strings.Compare(left.SourceID(), right.SourceID())
	})

	return normalizedItems
}
