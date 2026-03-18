package catalog_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkill(t *testing.T) {
	t.Run("normalize discovered skill metadata", func(t *testing.T) {
		got, err := catalog.NewSkill("anthropics-skills-skills-75224e3c", "reviewer/./", " Reviewer ", " Review pull requests carefully. ")

		require.NoError(t, err)
		assert.Equal(t, "anthropics-skills-skills-75224e3c:reviewer", got.ID())
		assert.Equal(t, "anthropics-skills-skills-75224e3c", got.SourceID())
		assert.Equal(t, "Reviewer", got.Name())
		assert.Equal(t, "Review pull requests carefully.", got.Description())
		assert.Equal(t, "reviewer", got.RelativePath())
		assert.Equal(t, "reviewer/SKILL.md", got.FilePath())
	})

	t.Run("resolve root skill file path", func(t *testing.T) {
		got, err := catalog.NewSkill("anthropics-skills-skills-75224e3c", "", "Root Skill", "")

		require.NoError(t, err)
		assert.Equal(t, "SKILL.md", got.FilePath())
	})

	t.Run("return error when source id is missing", func(t *testing.T) {
		_, err := catalog.NewSkill("", "reviewer", "Reviewer", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill source id required")
	})

	t.Run("return error when skill name is missing", func(t *testing.T) {
		_, err := catalog.NewSkill("anthropics-skills-skills-75224e3c", "reviewer", "", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill name required")
	})

	t.Run("return error when relative path escapes source subtree", func(t *testing.T) {
		_, err := catalog.NewSkill("anthropics-skills-skills-75224e3c", "../escape", "Reviewer", "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must stay within the source subtree")
	})
}
