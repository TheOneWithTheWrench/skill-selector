package profile

import (
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

const (
	// DefaultName is the built-in profile that always exists as the fallback selection owner.
	DefaultName = "Default"
)

// Default returns the built-in profile with an empty saved selection.
func Default() Profile {
	return Profile{name: DefaultName}
}

// New validates one profile name and normalizes its saved skill identities.
func New(name string, selected ...skill_identity.Identity) (Profile, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return Profile{}, fmt.Errorf("profile name required")
	}

	return Profile{
		name:     normalizedName,
		selected: skill_identity.NewIdentities(selected...),
	}, nil
}

// Profile keeps one named saved selection.
type Profile struct {
	name     string
	selected skill_identity.Identities
}

// Name returns the stable user-facing profile name.
func (p Profile) Name() string {
	return p.name
}

// Selected returns a copy of the saved selection owned by this profile.
func (p Profile) Selected() skill_identity.Identities {
	return append(skill_identity.Identities(nil), p.selected...)
}

// SelectedCount returns how many saved skill identities belong to this profile.
func (p Profile) SelectedCount() int {
	return len(p.selected)
}

// WithoutSource returns the same profile with every saved skill from one source removed.
func (p Profile) WithoutSource(sourceID string) Profile {
	filtered := make(skill_identity.Identities, 0, len(p.selected))
	for _, identity := range p.selected {
		if identity.SourceID() == sourceID {
			continue
		}

		filtered = append(filtered, identity)
	}

	return Profile{
		name:     p.name,
		selected: skill_identity.NewIdentities(filtered...),
	}
}

func normalizeName(name string) string {
	normalizedName := strings.TrimSpace(name)
	if strings.EqualFold(normalizedName, DefaultName) {
		return DefaultName
	}

	return normalizedName
}

func key(name string) string {
	return strings.ToLower(normalizeName(name))
}
