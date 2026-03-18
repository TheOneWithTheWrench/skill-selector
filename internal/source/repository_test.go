package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileRepository(t *testing.T) {
	t.Run("return error when path is empty", func(t *testing.T) {
		_, err := source.NewFileRepository("")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sources path required")
	})
}

func TestFileRepository(t *testing.T) {
	var (
		newRepository = func(t *testing.T) (*source.FileRepository, string) {
			path := filepath.Join(t.TempDir(), "sources.json")
			repository, err := source.NewFileRepository(path)
			require.NoError(t, err)
			return repository, path
		}
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
	)

	t.Run("load empty sources when file does not exist", func(t *testing.T) {
		repository, _ := newRepository(t)

		got, err := repository.Load()

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("save and load configured sources", func(t *testing.T) {
		var (
			repository, _  = newRepository(t)
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
		)

		err := repository.Save(source.NewSources(reviewerSource, rootSource))

		require.NoError(t, err)

		got, err := repository.Load()
		require.NoError(t, err)
		assert.Equal(t, source.Sources{rootSource, reviewerSource}, got)
	})

	t.Run("normalize duplicate entries from file", func(t *testing.T) {
		var (
			repository, path = newRepository(t)
			rootSource       = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource   = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
		)

		err := os.WriteFile(path, []byte(`{
  "version": 1,
  "sources": [
    {"url": "https://github.com/anthropics/skills/tree/main/skills/reviewer"},
    {"url": "https://github.com/anthropics/skills/tree/main/skills"},
    {"url": "https://github.com/anthropics/skills/tree/main/skills/reviewer"}
  ]
}
`), 0o644)
		require.NoError(t, err)

		got, err := repository.Load()
		require.NoError(t, err)
		assert.Equal(t, source.Sources{rootSource, reviewerSource}, got)
	})
}
