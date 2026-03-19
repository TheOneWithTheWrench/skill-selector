package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/fileutil"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
)

const repositoryVersion = 2

// Repository loads and saves persisted profiles.
type Repository interface {
	Load() (Profiles, error)
	Save(Profiles) error
}

// FileRepository persists profiles as a versioned JSON file.
type FileRepository struct {
	path string
}

type fileProfiles struct {
	Version       int           `json:"version"`
	ActiveProfile string        `json:"active_profile"`
	Profiles      []fileProfile `json:"profiles"`
}

type fileProfile struct {
	Name           string         `json:"name"`
	SelectedSkills []fileIdentity `json:"selected_skills,omitempty"`
}

type fileIdentity struct {
	SourceID     string `json:"source_id"`
	RelativePath string `json:"relative_path"`
}

type legacyProfiles struct {
	Version int         `json:"version"`
	Default fileProfile `json:"default"`
}

// NewFileRepository binds profile persistence to one JSON file path.
func NewFileRepository(path string) (*FileRepository, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("profiles path required")
	}

	return &FileRepository{path: path}, nil
}

// Load reads the stored profiles and falls back to the built-in default when the file does not exist.
func (r FileRepository) Load() (Profiles, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultProfiles(), nil
	}
	if err != nil {
		return Profiles{}, fmt.Errorf("read profiles file %q: %w", r.path, err)
	}

	var stored fileProfiles
	if err := json.Unmarshal(data, &stored); err == nil && isCurrentFile(stored) {
		return decodeProfilesFile(r.path, stored)
	}

	var legacy legacyProfiles
	if err := json.Unmarshal(data, &legacy); err != nil {
		return Profiles{}, fmt.Errorf("decode profiles file %q: %w", r.path, err)
	}

	return decodeLegacyProfilesFile(r.path, legacy)
}

// Save writes the normalized profiles using the current repository schema.
func (r FileRepository) Save(profiles Profiles) error {
	stored := fileProfiles{
		Version:       repositoryVersion,
		ActiveProfile: profiles.ActiveName(),
		Profiles:      make([]fileProfile, 0, len(profiles.All())),
	}

	for _, item := range profiles.All() {
		storedProfile := fileProfile{
			Name:           item.Name(),
			SelectedSkills: encodeIdentities(item.Selected()),
		}
		stored.Profiles = append(stored.Profiles, storedProfile)
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profiles file: %w", err)
	}

	if err := fileutil.WriteFile(r.path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write profiles file %q: %w", r.path, err)
	}

	return nil
}

func decodeProfilesFile(path string, stored fileProfiles) (Profiles, error) {
	profiles := make([]Profile, 0, len(stored.Profiles))
	for _, storedProfile := range stored.Profiles {
		decodedProfile, err := decodeProfile(storedProfile)
		if err != nil {
			return Profiles{}, fmt.Errorf("decode profiles file %q: %w", path, err)
		}

		profiles = append(profiles, decodedProfile)
	}

	return NewProfiles(stored.ActiveProfile, profiles...), nil
}

func decodeLegacyProfilesFile(path string, legacy legacyProfiles) (Profiles, error) {
	if legacy.Version == 0 && len(legacy.Default.SelectedSkills) == 0 {
		return DefaultProfiles(), nil
	}

	decodedDefault, err := decodeProfile(fileProfile{
		Name:           DefaultName,
		SelectedSkills: legacy.Default.SelectedSkills,
	})
	if err != nil {
		return Profiles{}, fmt.Errorf("decode profiles file %q: %w", path, err)
	}

	return NewProfiles(DefaultName, decodedDefault), nil
}

func decodeProfile(stored fileProfile) (Profile, error) {
	identities, err := decodeIdentities(stored.SelectedSkills)
	if err != nil {
		return Profile{}, err
	}

	return New(stored.Name, identities...)
}

func decodeIdentities(stored []fileIdentity) (skillidentity.Identities, error) {
	identities := make(skillidentity.Identities, 0, len(stored))
	for _, item := range stored {
		identity, err := skillidentity.New(item.SourceID, item.RelativePath)
		if err != nil {
			return nil, err
		}

		identities = append(identities, identity)
	}

	return skillidentity.NewIdentities(identities...), nil
}

func encodeIdentities(identities skillidentity.Identities) []fileIdentity {
	stored := make([]fileIdentity, 0, len(identities))
	for _, identity := range identities {
		stored = append(stored, fileIdentity{
			SourceID:     identity.SourceID(),
			RelativePath: identity.RelativePath(),
		})
	}

	return stored
}

func isCurrentFile(stored fileProfiles) bool {
	return stored.ActiveProfile != "" || stored.Profiles != nil
}
