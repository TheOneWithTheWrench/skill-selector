package source_test

import (
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirror(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
	)

	t.Run("build mirror paths from source", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			cloneRoot        = filepath.Join("/tmp", "skill-selector", "sources")
		)

		mirror, err := source.NewMirror(configuredSource, cloneRoot)

		require.NoError(t, err)
		assert.Equal(t, configuredSource.ID(), mirror.ID())
		assert.Equal(t, filepath.Join(cloneRoot, configuredSource.ID()), mirror.ClonePath)
		assert.Equal(t, filepath.Join(mirror.ClonePath, "skills"), mirror.SubtreePath())
		assert.Equal(t, filepath.Join(mirror.ClonePath, "skills", "reviewer"), mirror.SkillPath("reviewer"))
	})

	t.Run("keep skill path inside subtree", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			cloneRoot        = filepath.Join("/tmp", "skill-selector", "sources")
		)

		mirror, err := source.NewMirror(configuredSource, cloneRoot)

		require.NoError(t, err)
		assert.Equal(t, mirror.SubtreePath(), mirror.SkillPath("../escape"))
	})
}
