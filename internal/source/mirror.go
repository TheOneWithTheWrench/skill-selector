package source

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Mirror is the managed local clone of a Source. Clone and pull behavior belongs to a separate service.
type Mirror struct {
	Source    Source
	ClonePath string
}

// NewMirror resolves a source into its managed local clone location under cloneRoot.
func NewMirror(configuredSource Source, cloneRoot string) (Mirror, error) {
	if strings.TrimSpace(cloneRoot) == "" {
		return Mirror{}, fmt.Errorf("clone root required")
	}

	return Mirror{
		Source:    configuredSource,
		ClonePath: filepath.Join(cloneRoot, configuredSource.ID()),
	}, nil
}

// NewMirrors resolves a source collection into local mirrors under a shared clone root.
func NewMirrors(configuredSources Sources, cloneRoot string) ([]Mirror, error) {
	mirrors := make([]Mirror, 0, len(configuredSources))

	for _, configuredSource := range configuredSources {
		mirror, err := NewMirror(configuredSource, cloneRoot)
		if err != nil {
			return nil, err
		}

		mirrors = append(mirrors, mirror)
	}

	return mirrors, nil
}

// ID returns the stable source identifier for the mirrored source.
func (m Mirror) ID() string {
	return m.Source.ID()
}

// SubtreePath returns the local directory that should be scanned for skills.
func (m Mirror) SubtreePath() string {
	if m.Source.Subpath() == "" {
		return m.ClonePath
	}

	return filepath.Join(m.ClonePath, filepath.FromSlash(m.Source.Subpath()))
}

// SkillPath resolves a skill path within the mirrored subtree and prevents escaping that root.
func (m Mirror) SkillPath(relativePath string) string {
	if relativePath == "" || relativePath == "." {
		return m.SubtreePath()
	}

	return safeJoin(m.SubtreePath(), relativePath)
}

func safeJoin(rootPath string, relativePath string) string {
	cleaned := filepath.Clean(filepath.FromSlash(relativePath))
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return rootPath
	}

	joined := filepath.Join(rootPath, cleaned)
	cleanRoot := filepath.Clean(rootPath)
	rootPrefix := cleanRoot + string(filepath.Separator)
	if joined != cleanRoot && !strings.HasPrefix(joined, rootPrefix) {
		return cleanRoot
	}

	return joined
}
