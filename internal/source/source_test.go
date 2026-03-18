package source_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("parse github tree url", func(t *testing.T) {
		got, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")

		require.NoError(t, err)
		assert.Equal(t, "https://github.com/anthropics/skills/tree/main/skills", got.URL())
		assert.Equal(t, "anthropics", got.Owner())
		assert.Equal(t, "skills", got.Repo())
		assert.Equal(t, "main", got.Ref())
		assert.Equal(t, "skills", got.Subpath())
		assert.Equal(t, "anthropics/skills", got.RepoSlug())
	})

	t.Run("trim whitespace before parsing", func(t *testing.T) {
		got, err := source.Parse("  https://github.com/anthropics/skills/tree/main/skills  ")

		require.NoError(t, err)
		assert.Equal(t, "https://github.com/anthropics/skills/tree/main/skills", got.URL())
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

	t.Run("return error when url does not contain tree segment", func(t *testing.T) {
		_, err := source.Parse("https://github.com/anthropics/skills/blob/main/skills")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must contain /tree/")
	})
}

func TestSourceID(t *testing.T) {
	t.Run("return stable source id for repo ref and subtree", func(t *testing.T) {
		configuredSource, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")

		require.NoError(t, err)
		assert.Equal(t, "anthropics-skills-skills-75224e3c", configuredSource.ID())
	})
}
