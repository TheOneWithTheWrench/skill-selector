package catalog

import (
	"fmt"
	"path"
	"strings"
)

// Skill is one discovered skill directory within a source subtree.
type Skill struct {
	sourceID     string
	name         string
	description  string
	relativePath string
}

// NewSkill validates discovered skill metadata and normalizes its relative path.
func NewSkill(sourceID string, relativePath string, name string, description string) (Skill, error) {
	normalizedSourceID := strings.TrimSpace(sourceID)
	if normalizedSourceID == "" {
		return Skill{}, fmt.Errorf("skill source id required")
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return Skill{}, fmt.Errorf("skill name required")
	}

	normalizedRelativePath := normalizeRelativePath(relativePath)
	if normalizedRelativePath == ".." || strings.HasPrefix(normalizedRelativePath, "../") || strings.HasPrefix(normalizedRelativePath, "/") {
		return Skill{}, fmt.Errorf("skill relative path must stay within the source subtree: %q", relativePath)
	}

	return Skill{
		sourceID:     normalizedSourceID,
		name:         normalizedName,
		description:  strings.TrimSpace(description),
		relativePath: normalizedRelativePath,
	}, nil
}

// ID returns the stable identifier for the discovered skill within its source.
func (s Skill) ID() string {
	return s.sourceID + ":" + s.relativePath
}

// SourceID returns the source that produced the skill.
func (s Skill) SourceID() string {
	return s.sourceID
}

// Name returns the display name shown for the skill.
func (s Skill) Name() string {
	return s.name
}

// Description returns the one-line preview description for the skill.
func (s Skill) Description() string {
	return s.description
}

// RelativePath returns the skill directory path relative to the source subtree.
func (s Skill) RelativePath() string {
	return s.relativePath
}

// FilePath returns the relative path to the skill's `SKILL.md` file.
func (s Skill) FilePath() string {
	if s.relativePath == "" {
		return "SKILL.md"
	}

	return path.Join(s.relativePath, "SKILL.md")
}

func normalizeRelativePath(relativePath string) string {
	normalizedRelativePath := path.Clean(strings.TrimSpace(relativePath))
	if normalizedRelativePath == "." {
		return ""
	}

	return normalizedRelativePath
}
