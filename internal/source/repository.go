package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/fileutil"
)

const repositoryVersion = 1

// Repository loads and saves configured sources.
type Repository interface {
	Load() (Sources, error)
	Save(Sources) error
}

// FileRepository persists configured sources as a versioned JSON file.
type FileRepository struct {
	path string
}

type fileSources struct {
	Version int          `json:"version"`
	Sources []fileSource `json:"sources"`
}

type fileSource struct {
	URL string `json:"url"`
}

// NewFileRepository binds source persistence to a single file path.
func NewFileRepository(path string) (*FileRepository, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("sources path required")
	}

	return &FileRepository{path: path}, nil
}

// Load reads the source file and returns the normalized configured sources.
func (r FileRepository) Load() (Sources, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read sources file %q: %w", r.path, err)
	}

	var stored fileSources
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("decode sources file %q: %w", r.path, err)
	}

	configuredSources := make(Sources, 0, len(stored.Sources))
	for _, storedSource := range stored.Sources {
		configuredSource, err := Parse(storedSource.URL)
		if err != nil {
			return nil, fmt.Errorf("decode sources file %q: %w", r.path, err)
		}

		configuredSources = append(configuredSources, configuredSource)
	}

	return NewSources(configuredSources...), nil
}

// Save writes the configured sources to disk using the repository file schema.
func (r FileRepository) Save(configuredSources Sources) error {
	stored := fileSources{
		Version: repositoryVersion,
		Sources: make([]fileSource, 0, len(configuredSources)),
	}

	for _, configuredSource := range configuredSources {
		stored.Sources = append(stored.Sources, fileSource{URL: configuredSource.URL()})
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sources file: %w", err)
	}

	if err := fileutil.WriteFile(r.path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write sources file %q: %w", r.path, err)
	}

	return nil
}
