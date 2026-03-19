package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/fileutil"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
)

const repositoryVersion = 1

// Repository loads and saves catalog snapshots.
type Repository interface {
	Load() (Catalog, error)
	Save(Catalog) error
}

// FileRepository persists the catalog as a versioned JSON file.
type FileRepository struct {
	path string
}

type fileCatalog struct {
	Version     int         `json:"version"`
	GeneratedAt time.Time   `json:"generated_at"`
	Skills      []fileSkill `json:"skills"`
}

type fileSkill struct {
	SourceID     string `json:"source_id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	RelativePath string `json:"relative_path"`
	SkillFile    string `json:"skill_file"`
}

// NewFileRepository binds catalog persistence to a single file path.
func NewFileRepository(path string) (*FileRepository, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("catalog path required")
	}

	return &FileRepository{path: path}, nil
}

// Load reads the catalog file and returns the normalized catalog snapshot.
func (r FileRepository) Load() (Catalog, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return Catalog{}, nil
	}
	if err != nil {
		return Catalog{}, fmt.Errorf("read catalog file %q: %w", r.path, err)
	}

	var stored fileCatalog
	if err := json.Unmarshal(data, &stored); err != nil {
		return Catalog{}, fmt.Errorf("decode catalog file %q: %w", r.path, err)
	}

	discoveredSkills := make(Skills, 0, len(stored.Skills))
	for _, storedSkill := range stored.Skills {
		identity, err := skillidentity.New(storedSkill.SourceID, storedSkill.RelativePath)
		if err != nil {
			return Catalog{}, fmt.Errorf("decode catalog file %q: %w", r.path, err)
		}

		discoveredSkill, err := NewSkill(identity, storedSkill.Name, storedSkill.Description)
		if err != nil {
			return Catalog{}, fmt.Errorf("decode catalog file %q: %w", r.path, err)
		}

		discoveredSkills = append(discoveredSkills, discoveredSkill)
	}

	return NewCatalog(stored.GeneratedAt, discoveredSkills...), nil
}

// Save writes the catalog snapshot to disk using the repository file schema.
func (r FileRepository) Save(current Catalog) error {
	discoveredSkills := current.Skills()
	stored := fileCatalog{
		Version:     repositoryVersion,
		GeneratedAt: current.IndexedAt(),
		Skills:      make([]fileSkill, 0, len(discoveredSkills)),
	}

	for _, discoveredSkill := range discoveredSkills {
		stored.Skills = append(stored.Skills, fileSkill{
			SourceID:     discoveredSkill.SourceID(),
			Name:         discoveredSkill.Name(),
			Description:  discoveredSkill.Description(),
			RelativePath: discoveredSkill.RelativePath(),
			SkillFile:    discoveredSkill.FilePath(),
		})
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("encode catalog file: %w", err)
	}

	if err := fileutil.WriteFile(r.path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write catalog file %q: %w", r.path, err)
	}

	return nil
}
