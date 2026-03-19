package skillidentity

import (
	"fmt"
	"path"
	"strings"
)

// Identity identifies a skill by source and relative path without carrying catalog metadata.
type Identity struct {
	sourceID     string
	relativePath string
}

// New validates and normalizes a lightweight skill identity.
func New(sourceID string, relativePath string) (Identity, error) {
	normalizedSourceID := strings.TrimSpace(sourceID)
	if normalizedSourceID == "" {
		return Identity{}, fmt.Errorf("skill identity source id required")
	}

	normalizedRelativePath := path.Clean(strings.TrimSpace(relativePath))
	if normalizedRelativePath == "." {
		normalizedRelativePath = ""
	}
	if normalizedRelativePath == ".." || strings.HasPrefix(normalizedRelativePath, "../") || strings.HasPrefix(normalizedRelativePath, "/") {
		return Identity{}, fmt.Errorf("skill identity path must stay within the source subtree: %q", relativePath)
	}

	return Identity{
		sourceID:     normalizedSourceID,
		relativePath: normalizedRelativePath,
	}, nil
}

// SourceID returns the source that owns the referenced skill.
func (i Identity) SourceID() string {
	return i.sourceID
}

// RelativePath returns the skill path relative to the source subtree.
func (i Identity) RelativePath() string {
	return i.relativePath
}

// Key returns the stable string form used for deduplication and indexing.
func (i Identity) Key() string {
	return i.sourceID + ":" + i.relativePath
}
