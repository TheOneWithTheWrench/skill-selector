package skillidentity_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("normalize source id and relative path", func(t *testing.T) {
		got, err := skillidentity.New(" source-a ", "reviewer/./")

		require.NoError(t, err)
		assert.Equal(t, "source-a", got.SourceID())
		assert.Equal(t, "reviewer", got.RelativePath())
		assert.Equal(t, "source-a:reviewer", got.Key())
	})

	t.Run("allow root skill identity", func(t *testing.T) {
		got, err := skillidentity.New("source-a", "")

		require.NoError(t, err)
		assert.Equal(t, "", got.RelativePath())
		assert.Equal(t, "source-a:", got.Key())
	})

	t.Run("return error when source id is missing", func(t *testing.T) {
		_, err := skillidentity.New("", "reviewer")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source id required")
	})

	t.Run("return error when path escapes source subtree", func(t *testing.T) {
		_, err := skillidentity.New("source-a", "../escape")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must stay within the source subtree")
	})
}
