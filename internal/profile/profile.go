package profile

import (
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
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
func New(name string, selected ...skillidentity.Identity) (Profile, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return Profile{}, fmt.Errorf("profile name required")
	}

	return Profile{
		name:     normalizedName,
		selected: skillidentity.NewIdentities(selected...),
	}, nil
}

// Profile keeps one named saved selection.
type Profile struct {
	name     string
	selected skillidentity.Identities
}

// Name returns the stable user-facing profile name.
func (p Profile) Name() string {
	return p.name
}

// Selected returns a copy of the saved selection owned by this profile.
func (p Profile) Selected() skillidentity.Identities {
	return append(skillidentity.Identities(nil), p.selected...)
}

// SelectedCount returns how many saved skill identities belong to this profile.
func (p Profile) SelectedCount() int {
	return len(p.selected)
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
