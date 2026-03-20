package catalog_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkill(t *testing.T) {
	newIdentity := func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
		t.Helper()

		identity, err := skill_identity.New(sourceID, relativePath)
		require.NoError(t, err)
		return identity
	}

	t.Run("normalize discovered skill metadata", func(t *testing.T) {
		got, err := catalog.NewSkill(newIdentity(t, "anthropics-skills-skills-75224e3c", "reviewer/./"), " Reviewer ", " Review pull requests carefully. ", " code-review ", "CODE-REVIEW", "security")

		require.NoError(t, err)
		assert.Equal(t, "anthropics-skills-skills-75224e3c:reviewer", got.ID())
		assert.Equal(t, "anthropics-skills-skills-75224e3c:reviewer", got.Identity().Key())
		assert.Equal(t, "anthropics-skills-skills-75224e3c", got.SourceID())
		assert.Equal(t, "Reviewer", got.Name())
		assert.Equal(t, "Review pull requests carefully.", got.Description())
		assert.Equal(t, []string{"code-review", "security"}, got.Tags())
		assert.Equal(t, "reviewer", got.RelativePath())
		assert.Equal(t, "reviewer/SKILL.md", got.FilePath())
	})

	t.Run("resolve root skill file path", func(t *testing.T) {
		got, err := catalog.NewSkill(newIdentity(t, "anthropics-skills-skills-75224e3c", ""), "Root Skill", "")

		require.NoError(t, err)
		assert.Equal(t, "SKILL.md", got.FilePath())
	})

	t.Run("return error when skill name is missing", func(t *testing.T) {
		_, err := catalog.NewSkill(newIdentity(t, "anthropics-skills-skills-75224e3c", "reviewer"), "", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill name required")
	})
}
