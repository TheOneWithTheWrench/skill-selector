package profile

import (
	"fmt"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

// DefaultProfiles returns the normalized persisted state used before any profile file exists.
func DefaultProfiles() Profiles {
	return NewProfiles(DefaultName, Default())
}

// NewProfiles normalizes a set of profiles, guarantees the default profile exists,
// and falls back to the default profile when the requested active profile is missing.
func NewProfiles(activeName string, items ...Profile) Profiles {
	var (
		defaultProfile = Default()
		profiles       = make([]Profile, 0, len(items)+1)
		seen           = make(map[string]struct{}, len(items)+1)
	)

	for _, item := range items {
		normalizedName := normalizeName(item.name)
		if normalizedName == "" {
			continue
		}

		normalized := Profile{
			name:     normalizedName,
			selected: skill_identity.NewIdentities(item.selected...),
		}

		profileKey := key(normalizedName)
		if _, ok := seen[profileKey]; ok {
			continue
		}

		seen[profileKey] = struct{}{}
		if normalizedName == DefaultName {
			defaultProfile = normalized
			continue
		}

		profiles = append(profiles, normalized)
	}

	slices.SortFunc(profiles, func(left Profile, right Profile) int {
		return strings.Compare(left.Name(), right.Name())
	})

	profiles = append([]Profile{defaultProfile}, profiles...)

	normalizedActive := normalizeName(activeName)
	if normalizedActive == "" || !containsProfile(profiles, normalizedActive) {
		normalizedActive = DefaultName
	}

	return Profiles{
		activeName: normalizedActive,
		items:      profiles,
	}
}

// Profiles keeps the saved selection state across every named profile together with the active one.
type Profiles struct {
	activeName string
	items      []Profile
}

// Active returns the currently active profile.
func (p Profiles) Active() Profile {
	for _, item := range p.items {
		if item.Name() == p.activeName {
			return item
		}
	}

	return Default()
}

// ActiveName returns the name of the profile that currently owns the saved selection.
func (p Profiles) ActiveName() string {
	if p.activeName == "" {
		return DefaultName
	}

	return p.activeName
}

// All returns the normalized profiles in stable display order.
func (p Profiles) All() []Profile {
	return append([]Profile(nil), p.items...)
}

// Find returns one profile by name using the same normalization rules as persistence.
func (p Profiles) Find(name string) (Profile, bool) {
	targetKey := key(name)
	for _, item := range p.items {
		if key(item.Name()) == targetKey {
			return item, true
		}
	}

	return Profile{}, false
}

// Create adds a new empty profile while preserving the current active profile.
func (p Profiles) Create(name string) (Profiles, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return Profiles{}, fmt.Errorf("profile name required")
	}
	if _, ok := p.Find(normalizedName); ok {
		return Profiles{}, fmt.Errorf("profile already exists: %s", normalizedName)
	}

	created, err := New(normalizedName)
	if err != nil {
		return Profiles{}, err
	}

	return NewProfiles(p.ActiveName(), append(p.All(), created)...), nil
}

// Rename changes one profile name and updates the active profile name when needed.
func (p Profiles) Rename(currentName string, newName string) (Profiles, error) {
	normalizedCurrentName := normalizeName(currentName)
	normalizedNewName := normalizeName(newName)
	if normalizedCurrentName == "" || normalizedNewName == "" {
		return Profiles{}, fmt.Errorf("profile names required")
	}
	if normalizedCurrentName == DefaultName {
		return Profiles{}, fmt.Errorf("cannot rename %s profile", DefaultName)
	}

	currentIndex := p.index(normalizedCurrentName)
	if currentIndex == -1 {
		return Profiles{}, fmt.Errorf("profile not found: %s", normalizedCurrentName)
	}
	if key(normalizedCurrentName) != key(normalizedNewName) {
		if _, ok := p.Find(normalizedNewName); ok {
			return Profiles{}, fmt.Errorf("profile already exists: %s", normalizedNewName)
		}
	}

	nextProfiles := p.All()
	nextProfiles[currentIndex].name = normalizedNewName

	activeName := p.ActiveName()
	if key(activeName) == key(normalizedCurrentName) {
		activeName = normalizedNewName
	}

	return NewProfiles(activeName, nextProfiles...), nil
}

// Remove deletes one inactive non-default profile.
func (p Profiles) Remove(name string) (Profiles, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return Profiles{}, fmt.Errorf("profile name required")
	}
	if normalizedName == DefaultName {
		return Profiles{}, fmt.Errorf("cannot remove %s profile", DefaultName)
	}
	if key(p.ActiveName()) == key(normalizedName) {
		return Profiles{}, fmt.Errorf("cannot remove active profile: %s", normalizedName)
	}

	removeIndex := p.index(normalizedName)
	if removeIndex == -1 {
		return Profiles{}, fmt.Errorf("profile not found: %s", normalizedName)
	}

	nextProfiles := append(append([]Profile(nil), p.items[:removeIndex]...), p.items[removeIndex+1:]...)
	return NewProfiles(p.ActiveName(), nextProfiles...), nil
}

// SetActiveSelection replaces the saved selection owned by the active profile.
func (p Profiles) SetActiveSelection(selected skill_identity.Identities) (Profiles, error) {
	return p.SetSelection(p.ActiveName(), selected)
}

// SetSelection replaces the saved selection owned by one named profile.
func (p Profiles) SetSelection(name string, selected skill_identity.Identities) (Profiles, error) {
	profileIndex := p.index(name)
	if profileIndex == -1 {
		return Profiles{}, fmt.Errorf("profile not found: %s", normalizeName(name))
	}

	nextProfiles := p.All()
	nextProfiles[profileIndex].selected = skill_identity.NewIdentities(selected...)

	return NewProfiles(p.ActiveName(), nextProfiles...), nil
}

// Switch changes which profile owns the saved selection without syncing it automatically.
func (p Profiles) Switch(name string) (Profiles, error) {
	normalizedName := normalizeName(name)
	if normalizedName == "" {
		return Profiles{}, fmt.Errorf("profile name required")
	}
	if _, ok := p.Find(normalizedName); !ok {
		return Profiles{}, fmt.Errorf("profile not found: %s", normalizedName)
	}

	return NewProfiles(normalizedName, p.All()...), nil
}

// WithoutSource removes every saved skill owned by one source across all profiles.
func (p Profiles) WithoutSource(sourceID string) Profiles {
	nextProfiles := make([]Profile, 0, len(p.items))
	for _, item := range p.items {
		nextProfiles = append(nextProfiles, item.WithoutSource(sourceID))
	}

	return NewProfiles(p.ActiveName(), nextProfiles...)
}

func (p Profiles) index(name string) int {
	targetKey := key(name)
	for index, item := range p.items {
		if key(item.Name()) == targetKey {
			return index
		}
	}

	return -1
}

func containsProfile(items []Profile, name string) bool {
	targetKey := key(name)
	for _, item := range items {
		if key(item.Name()) == targetKey {
			return true
		}
	}

	return false
}
