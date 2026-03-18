package skillref

import (
	"fmt"
	"path"
	"strings"
)

// Ref identifies a skill by source and relative path without carrying catalog metadata.
type Ref struct {
	sourceID     string
	relativePath string
}

// New validates and normalizes a lightweight skill reference.
func New(sourceID string, relativePath string) (Ref, error) {
	normalizedSourceID := strings.TrimSpace(sourceID)
	if normalizedSourceID == "" {
		return Ref{}, fmt.Errorf("skill ref source id required")
	}

	normalizedRelativePath := path.Clean(strings.TrimSpace(relativePath))
	if normalizedRelativePath == "." {
		normalizedRelativePath = ""
	}
	if normalizedRelativePath == ".." || strings.HasPrefix(normalizedRelativePath, "../") || strings.HasPrefix(normalizedRelativePath, "/") {
		return Ref{}, fmt.Errorf("skill ref path must stay within the source subtree: %q", relativePath)
	}

	return Ref{
		sourceID:     normalizedSourceID,
		relativePath: normalizedRelativePath,
	}, nil
}

// SourceID returns the source that owns the referenced skill.
func (r Ref) SourceID() string {
	return r.sourceID
}

// RelativePath returns the skill path relative to the source subtree.
func (r Ref) RelativePath() string {
	return r.relativePath
}

// Key returns the stable string form used for deduplication and indexing.
func (r Ref) Key() string {
	return r.sourceID + ":" + r.relativePath
}
