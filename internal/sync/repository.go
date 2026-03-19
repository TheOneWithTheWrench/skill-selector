package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/file_util"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

const repositoryVersion = 1

// ManifestRepository loads and saves sync manifests.
type ManifestRepository interface {
	LoadAll() ([]Manifest, error)
	Save(Manifest) error
}

// DirectoryManifestRepository persists one manifest file per adapter in a directory.
type DirectoryManifestRepository struct {
	dir string
}

type manifestFile struct {
	Version  int                `json:"version"`
	Adapter  string             `json:"adapter,omitempty"`
	Agent    string             `json:"agent,omitempty"`
	RootPath string             `json:"root_path,omitempty"`
	Skills   []manifestSkillRef `json:"skills"`
}

type manifestSkillRef struct {
	SourceID     string `json:"source_id"`
	RelativePath string `json:"relative_path"`
}

// NewDirectoryManifestRepository binds manifest persistence to one directory.
func NewDirectoryManifestRepository(dir string) (*DirectoryManifestRepository, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("sync manifest directory required")
	}

	return &DirectoryManifestRepository{dir: strings.TrimSpace(dir)}, nil
}

// LoadAll loads every manifest file in the directory and sorts them by adapter.
func (r DirectoryManifestRepository) LoadAll() ([]Manifest, error) {
	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read sync state directory %q: %w", r.dir, err)
	}

	var manifests []Manifest
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		manifest, err := r.load(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		manifests = append(manifests, manifest)
	}

	slices.SortFunc(manifests, func(left Manifest, right Manifest) int {
		return strings.Compare(left.Adapter(), right.Adapter())
	})

	return manifests, nil
}

// Save persists one manifest file using the adapter name as the filename.
func (r DirectoryManifestRepository) Save(manifest Manifest) error {
	if manifest.Adapter() == "" {
		return fmt.Errorf("manifest adapter required")
	}

	encoded := manifestFile{
		Version:  repositoryVersion,
		Adapter:  manifest.Adapter(),
		RootPath: manifest.RootPath(),
		Skills:   make([]manifestSkillRef, 0, len(manifest.Identities())),
	}

	for _, identity := range manifest.Identities() {
		encoded.Skills = append(encoded.Skills, manifestSkillRef{
			SourceID:     identity.SourceID(),
			RelativePath: identity.RelativePath(),
		})
	}

	data, err := json.MarshalIndent(encoded, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sync manifest for %q: %w", manifest.Adapter(), err)
	}

	path := filepath.Join(r.dir, manifest.Adapter()+".json")
	if err := file_util.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write sync manifest %q: %w", path, err)
	}

	return nil
}

func (r DirectoryManifestRepository) load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read sync manifest %q: %w", path, err)
	}

	var decoded manifestFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Manifest{}, fmt.Errorf("decode sync manifest %q: %w", path, err)
	}

	adapter := strings.TrimSpace(decoded.Adapter)
	if adapter == "" {
		adapter = strings.TrimSpace(decoded.Agent)
	}

	identities := make(skill_identity.Identities, 0, len(decoded.Skills))
	for _, item := range decoded.Skills {
		identity, err := skill_identity.New(item.SourceID, item.RelativePath)
		if err != nil {
			return Manifest{}, fmt.Errorf("decode sync manifest %q: %w", path, err)
		}

		identities = append(identities, identity)
	}

	manifest, err := NewManifest(adapter, decoded.RootPath, identities...)
	if err != nil {
		return Manifest{}, fmt.Errorf("decode sync manifest %q: %w", path, err)
	}

	return manifest, nil
}
