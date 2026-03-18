package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillref"
)

// Manifest is the persisted set of skill refs currently owned by one sync target.
type Manifest struct {
	adapter  string
	rootPath string
	refs     skillref.Refs
}

// NewManifest constructs normalized sync ownership state for one target adapter.
func NewManifest(adapter string, rootPath string, refs ...skillref.Ref) (Manifest, error) {
	normalizedAdapter := strings.TrimSpace(adapter)
	if normalizedAdapter == "" {
		return Manifest{}, fmt.Errorf("manifest adapter required")
	}

	normalizedRootPath := strings.TrimSpace(rootPath)
	if normalizedRootPath != "" {
		normalizedRootPath = filepath.Clean(normalizedRootPath)
	}

	return Manifest{
		adapter:  normalizedAdapter,
		rootPath: normalizedRootPath,
		refs:     skillref.NewRefs(refs...),
	}, nil
}

// Adapter returns the adapter that owns the manifest file.
func (m Manifest) Adapter() string {
	return m.adapter
}

// RootPath returns the filesystem root synced for the target.
func (m Manifest) RootPath() string {
	return m.rootPath
}

// Refs returns the normalized skill refs currently owned by the target.
func (m Manifest) Refs() skillref.Refs {
	return append(skillref.Refs(nil), m.refs...)
}

func (m Manifest) withRootPath(rootPath string) Manifest {
	normalizedRootPath := strings.TrimSpace(rootPath)
	if normalizedRootPath != "" {
		normalizedRootPath = filepath.Clean(normalizedRootPath)
	}

	m.rootPath = normalizedRootPath
	return m
}

func (m Manifest) withRefs(refs skillref.Refs) Manifest {
	m.refs = skillref.NewRefs(refs...)
	return m
}
