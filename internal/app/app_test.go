package app_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repository struct {
	loadResult source.Sources
	loadErr    error
	saveErr    error
	saveCalls  []source.Sources
}

func (r *repository) Load() (source.Sources, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}

	return r.loadResult, nil
}

func (r *repository) Save(configuredSources source.Sources) error {
	r.saveCalls = append(r.saveCalls, configuredSources)
	if r.saveErr != nil {
		return r.saveErr
	}

	r.loadResult = configuredSources
	return nil
}

func TestNew(t *testing.T) {
	t.Run("return error when with source repository option has nil repository", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithSourceRepository(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source repository required")
	})

	t.Run("create app with default source repository", func(t *testing.T) {
		sut, err := app.New(testRuntime(t))

		require.NoError(t, err)
		require.NotNil(t, sut)
	})
}

func TestSources(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newSut = func(t *testing.T, repository *repository) *app.App {
			sut, err := app.New(testRuntime(t), app.WithSourceRepository(repository))
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("list sources", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &repository{loadResult: source.NewSources(configuredSource)}
			sut              = newSut(t, repository)
		)

		got, err := sut.ListSources()

		require.NoError(t, err)
		assert.Equal(t, source.Sources{configuredSource}, got)
	})

	t.Run("add and list sources", func(t *testing.T) {
		var (
			url              = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, url)
			repository       = &repository{loadResult: source.NewSources()}
			sut              = newSut(t, repository)
		)

		configuredSources, addedSource, err := sut.AddSource(url)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, addedSource)
		assert.Equal(t, source.Sources{configuredSource}, configuredSources)
		require.Len(t, repository.saveCalls, 1)
		assert.Equal(t, source.Sources{configuredSource}, repository.saveCalls[0])

		loadedSources, err := sut.ListSources()
		require.NoError(t, err)
		assert.Equal(t, configuredSources, loadedSources)
	})

	t.Run("remove source", func(t *testing.T) {
		var (
			url              = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, url)
			repository       = &repository{loadResult: source.NewSources(configuredSource)}
			sut              = newSut(t, repository)
		)

		configuredSources, removedSource, err := sut.RemoveSource(url)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, removedSource)
		assert.Nil(t, configuredSources)
		require.Len(t, repository.saveCalls, 1)
		assert.Nil(t, repository.saveCalls[0])
	})

	t.Run("return repository load error while adding source", func(t *testing.T) {
		var (
			expectedErr = errors.New("load failed")
			repository  = &repository{loadErr: expectedErr}
			sut         = newSut(t, repository)
		)

		_, _, err := sut.AddSource("https://github.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("return repository save error while removing source", func(t *testing.T) {
		var (
			expectedErr      = errors.New("save failed")
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &repository{
				loadResult: source.NewSources(configuredSource),
				saveErr:    expectedErr,
			}
			sut = newSut(t, repository)
		)

		_, _, err := sut.RemoveSource(configuredSource.URL())

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func testRuntime(t *testing.T) paths.Runtime {
	t.Helper()

	rootDir := t.TempDir()

	return paths.Runtime{
		CacheDir:     filepath.Join(rootDir, "cache"),
		DataDir:      filepath.Join(rootDir, "data"),
		SourcesFile:  filepath.Join(rootDir, "data", "sources.json"),
		SourcesDir:   filepath.Join(rootDir, "data", "sources"),
		CatalogFile:  filepath.Join(rootDir, "cache", "catalog.json"),
		ProfilesFile: filepath.Join(rootDir, "data", "profiles.json"),
		SyncStateDir: filepath.Join(rootDir, "data", "activations"),
		LogsDir:      filepath.Join(rootDir, "cache", "logs"),
	}
}
