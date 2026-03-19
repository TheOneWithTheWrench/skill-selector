package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
)

// Manifest is the persisted set of skill identities currently owned by one sync target.
type Manifest struct {
	adapter    string
	rootPath   string
	identities skillidentity.Identities
}

// NewManifest constructs normalized sync ownership state for one target adapter.
func NewManifest(adapter string, rootPath string, identities ...skillidentity.Identity) (Manifest, error) {
	normalizedAdapter := strings.TrimSpace(adapter)
	if normalizedAdapter == "" {
		return Manifest{}, fmt.Errorf("manifest adapter required")
	}

	normalizedRootPath := strings.TrimSpace(rootPath)
	if normalizedRootPath != "" {
		normalizedRootPath = filepath.Clean(normalizedRootPath)
	}

	return Manifest{
		adapter:    normalizedAdapter,
		rootPath:   normalizedRootPath,
		identities: skillidentity.NewIdentities(identities...),
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

// Identities returns the normalized skill identities currently owned by the target.
func (m Manifest) Identities() skillidentity.Identities {
	return append(skillidentity.Identities(nil), m.identities...)
}

func (m Manifest) withRootPath(rootPath string) Manifest {
	normalizedRootPath := strings.TrimSpace(rootPath)
	if normalizedRootPath != "" {
		normalizedRootPath = filepath.Clean(normalizedRootPath)
	}

	m.rootPath = normalizedRootPath
	return m
}

func (m Manifest) withIdentities(identities skillidentity.Identities) Manifest {
	m.identities = skillidentity.NewIdentities(identities...)
	return m
}
