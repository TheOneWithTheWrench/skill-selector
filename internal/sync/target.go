package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillref"
)

// Target is one sync destination with a root path and link mapping strategy.
type Target struct {
	adapter  string
	rootPath string
	linkPath func(skillref.Ref) string
}

// NewTarget validates a target before it participates in sync reconciliation.
func NewTarget(adapter string, rootPath string, linkPath func(skillref.Ref) string) (Target, error) {
	normalizedAdapter := strings.TrimSpace(adapter)
	if normalizedAdapter == "" {
		return Target{}, fmt.Errorf("target adapter required")
	}

	normalizedRootPath := strings.TrimSpace(rootPath)
	if normalizedRootPath == "" {
		return Target{}, fmt.Errorf("target root path required for %q", normalizedAdapter)
	}

	if linkPath == nil {
		return Target{}, fmt.Errorf("target link path resolver required for %q", normalizedAdapter)
	}

	return Target{
		adapter:  normalizedAdapter,
		rootPath: filepath.Clean(normalizedRootPath),
		linkPath: linkPath,
	}, nil
}

// Adapter returns the adapter name used for this target.
func (t Target) Adapter() string {
	return t.adapter
}

// RootPath returns the root directory that receives synced links.
func (t Target) RootPath() string {
	return t.rootPath
}

// LinkPath returns the destination path for one skill ref on the target.
func (t Target) LinkPath(ref skillref.Ref) string {
	return t.linkPath(ref)
}
