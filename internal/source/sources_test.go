package source_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSources(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
	)

	t.Run("normalize by source id and stable order", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
		)

		got := source.NewSources(reviewerSource, rootSource, reviewerSource)

		assert.Equal(t, source.Sources{rootSource, reviewerSource}, got)
	})
}

func TestSourcesAdd(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
	)

	t.Run("add source to collection", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			sut            = source.NewSources(rootSource)
		)

		got, err := sut.Add(reviewerSource)

		require.NoError(t, err)
		assert.Equal(t, source.Sources{rootSource, reviewerSource}, got)
	})

	t.Run("return error when source id already exists", func(t *testing.T) {
		var (
			rootSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			sut        = source.NewSources(rootSource)
		)

		_, err := sut.Add(rootSource)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source already exists")
	})
}

func TestSourcesRemove(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
	)

	t.Run("remove source by url", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			sut            = source.NewSources(rootSource, reviewerSource)
		)

		got, removedSource, err := sut.Remove(rootSource.Locator())

		require.NoError(t, err)
		assert.Equal(t, rootSource, removedSource)
		assert.Equal(t, source.Sources{reviewerSource}, got)
	})

	t.Run("remove source by source id", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			sut            = source.NewSources(rootSource, reviewerSource)
		)

		got, removedSource, err := sut.Remove(reviewerSource.ID())

		require.NoError(t, err)
		assert.Equal(t, reviewerSource, removedSource)
		assert.Equal(t, source.Sources{rootSource}, got)
	})

	t.Run("return error when source identifier is missing", func(t *testing.T) {
		sut := source.NewSources()

		_, _, err := sut.Remove("")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source identifier required")
	})

	t.Run("return error when source is not found", func(t *testing.T) {
		sut := source.NewSources()

		_, _, err := sut.Remove("missing")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source not found")
	})
}
