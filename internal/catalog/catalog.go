package catalog

import (
	"fmt"
	"strings"
	"time"
)

// Catalog is the current indexed inventory of discovered skills.
type Catalog struct {
	indexedAt time.Time
	skills    Skills
}

// NewCatalog normalizes discovered skills and records when the catalog was indexed.
func NewCatalog(indexedAt time.Time, skills ...Skill) Catalog {
	return Catalog{
		indexedAt: normalizeTime(indexedAt),
		skills:    NewSkills(skills...),
	}
}

// IndexedAt returns when the catalog snapshot was generated.
func (c Catalog) IndexedAt() time.Time {
	return c.indexedAt
}

// Skills returns a copy of the discovered skills in stable order.
func (c Catalog) Skills() Skills {
	return append(Skills(nil), c.skills...)
}

// ReplaceSource swaps the catalog entries for one source with a replacement skill set.
func (c Catalog) ReplaceSource(sourceID string, replacement Skills) (Catalog, error) {
	normalizedSourceID := strings.TrimSpace(sourceID)
	if normalizedSourceID == "" {
		return Catalog{}, fmt.Errorf("catalog source id required")
	}

	retainedSkills := make(Skills, 0, len(c.skills)+len(replacement))
	for _, skill := range c.skills {
		if skill.SourceID() != normalizedSourceID {
			retainedSkills = append(retainedSkills, skill)
		}
	}

	for _, skill := range replacement {
		if skill.SourceID() != normalizedSourceID {
			return Catalog{}, fmt.Errorf("catalog replacement skill source mismatch: %s", skill.ID())
		}

		retainedSkills = append(retainedSkills, skill)
	}

	return NewCatalog(c.indexedAt, retainedSkills...), nil
}

// RemoveSource removes all discovered skills belonging to one source.
func (c Catalog) RemoveSource(sourceID string) (Catalog, error) {
	return c.ReplaceSource(sourceID, nil)
}

func normalizeTime(indexedAt time.Time) time.Time {
	if indexedAt.IsZero() {
		return time.Time{}
	}

	return indexedAt.UTC()
}
