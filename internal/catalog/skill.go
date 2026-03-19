package catalog

import (
	"fmt"
	"path"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
)

// Skill combines a stable skill identity with discovered catalog metadata.
type Skill struct {
	identity    skillidentity.Identity
	name        string
	description string
}

// NewSkill validates discovered skill metadata for a known skill identity.
func NewSkill(identity skillidentity.Identity, name string, description string) (Skill, error) {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return Skill{}, fmt.Errorf("skill name required")
	}

	return Skill{
		identity:    identity,
		name:        normalizedName,
		description: strings.TrimSpace(description),
	}, nil
}

// ID returns the stable identifier for the discovered skill within its source.
func (s Skill) ID() string {
	return s.identity.Key()
}

// Identity returns the stable identity used to persist and reference the skill.
func (s Skill) Identity() skillidentity.Identity {
	return s.identity
}

// SourceID returns the source that produced the skill.
func (s Skill) SourceID() string {
	return s.identity.SourceID()
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
	return s.identity.RelativePath()
}

// FilePath returns the relative path to the skill's `SKILL.md` file.
func (s Skill) FilePath() string {
	if s.identity.RelativePath() == "" {
		return "SKILL.md"
	}

	return path.Join(s.identity.RelativePath(), "SKILL.md")
}
