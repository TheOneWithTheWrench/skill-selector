package source_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("parse github tree url", func(t *testing.T) {
		got, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")

		require.NoError(t, err)
		assert.Equal(t, "https://github.com/anthropics/skills/tree/main/skills", got.Locator())
		assert.Equal(t, "https://github.com/anthropics/skills.git", got.CloneURL())
		assert.Equal(t, "main", got.Ref())
		assert.Equal(t, "skills", got.Subpath())
	})

	t.Run("parse github repo root url", func(t *testing.T) {
		got, err := source.Parse("https://github.com/ComposioHQ/awesome-claude-skills")

		require.NoError(t, err)
		assert.Equal(t, "https://github.com/ComposioHQ/awesome-claude-skills", got.Locator())
		assert.Equal(t, "https://github.com/ComposioHQ/awesome-claude-skills.git", got.CloneURL())
		assert.Empty(t, got.Ref())
		assert.Empty(t, got.Subpath())
	})

	t.Run("trim whitespace before parsing", func(t *testing.T) {
		got, err := source.Parse("  https://github.com/anthropics/skills/tree/main/skills  ")

		require.NoError(t, err)
		assert.Equal(t, "https://github.com/anthropics/skills/tree/main/skills", got.Locator())
	})

	t.Run("clean subtree path", func(t *testing.T) {
		got, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills/./")

		require.NoError(t, err)
		assert.Equal(t, "skills", got.Subpath())
	})

	t.Run("return error when url is empty", func(t *testing.T) {
		_, err := source.Parse("")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source url required")
	})

	t.Run("return error when scheme is not https", func(t *testing.T) {
		_, err := source.Parse("http://github.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must use https")
	})

	t.Run("return error when host is not github", func(t *testing.T) {
		_, err := source.Parse("https://gitlab.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must point at github.com")
	})

	t.Run("return error when url is neither a repo nor tree url", func(t *testing.T) {
		_, err := source.Parse("https://github.com/anthropics/skills/blob/main/skills")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a GitHub repo or tree url")
	})
}

func TestSourceID(t *testing.T) {
	t.Run("return stable source id for repo ref and subtree", func(t *testing.T) {
		configuredSource, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")

		require.NoError(t, err)
		assert.Equal(t, "anthropics-skills-skills-75224e3c", configuredSource.ID())
	})

	t.Run("return stable source id for repo root without explicit ref", func(t *testing.T) {
		configuredSource, err := source.Parse("https://github.com/ComposioHQ/awesome-claude-skills")

		require.NoError(t, err)
		assert.Equal(t, "composiohq-awesome-claude-skills-d4a5ef49", configuredSource.ID())
	})
}
