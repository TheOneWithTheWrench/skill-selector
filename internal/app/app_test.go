package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sourceRepository struct {
	loadResult source.Sources
	loadErr    error
	saveErr    error
	saveCalls  []source.Sources
}

func (r *sourceRepository) Load() (source.Sources, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}

	return r.loadResult, nil
}

func (r *sourceRepository) Save(configuredSources source.Sources) error {
	r.saveCalls = append(r.saveCalls, configuredSources)
	if r.saveErr != nil {
		return r.saveErr
	}

	r.loadResult = configuredSources
	return nil
}

type sourceRefresher struct {
	refreshFunc func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error)
	calls       []source.Mirror
}

func (r *sourceRefresher) Refresh(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
	r.calls = append(r.calls, mirror)
	if r.refreshFunc != nil {
		return r.refreshFunc(ctx, mirror)
	}

	return source.RefreshResult{Mirror: mirror}, nil
}

type catalogRepository struct {
	saveErr   error
	saveCalls []catalog.Catalog
}

func (r *catalogRepository) Load() (catalog.Catalog, error) {
	return catalog.Catalog{}, nil
}

func (r *catalogRepository) Save(current catalog.Catalog) error {
	r.saveCalls = append(r.saveCalls, current)
	if r.saveErr != nil {
		return r.saveErr
	}

	return nil
}

type clock struct {
	now time.Time
}

func (c clock) Now() time.Time {
	return c.now
}

func TestNew(t *testing.T) {
	t.Run("return error when with source repository option has nil repository", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithSourceRepository(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source repository required")
	})

	t.Run("return error when with source refresher option has nil refresher", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithSourceRefresher(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source refresher required")
	})

	t.Run("return error when with catalog repository option has nil repository", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithCatalogRepository(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog repository required")
	})

	t.Run("return error when with catalog scanner option has nil scanner", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithCatalogScanner(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog scanner required")
	})

	t.Run("return error when with clock option has nil clock", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithClock(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "clock required")
	})

	t.Run("return error when with sync manifest repository option has nil repository", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithSyncManifestRepository(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sync manifest repository required")
	})

	t.Run("return error when with sync targets loader option has nil loader", func(t *testing.T) {
		_, err := app.New(paths.Runtime{}, app.WithSyncTargetsLoader(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sync targets loader required")
	})

	t.Run("create app with default dependencies", func(t *testing.T) {
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
		newSut = func(t *testing.T, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(testRuntime(t), optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("list sources", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &sourceRepository{loadResult: source.NewSources(configuredSource)}
			sut              = newSut(t, app.WithSourceRepository(repository))
		)

		got, err := sut.ListSources()

		require.NoError(t, err)
		assert.Equal(t, source.Sources{configuredSource}, got)
	})

	t.Run("add and list sources", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			repository       = &sourceRepository{loadResult: source.NewSources()}
			sut              = newSut(t, app.WithSourceRepository(repository))
		)

		configuredSources, addedSource, err := sut.AddSource(locator)

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
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			repository       = &sourceRepository{loadResult: source.NewSources(configuredSource)}
			sut              = newSut(t, app.WithSourceRepository(repository))
		)

		configuredSources, removedSource, err := sut.RemoveSource(locator)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, removedSource)
		assert.Nil(t, configuredSources)
		require.Len(t, repository.saveCalls, 1)
		assert.Nil(t, repository.saveCalls[0])
	})

	t.Run("return repository load error while adding source", func(t *testing.T) {
		var (
			expectedErr = errors.New("load failed")
			repository  = &sourceRepository{loadErr: expectedErr}
			sut         = newSut(t, app.WithSourceRepository(repository))
		)

		_, _, err := sut.AddSource("https://github.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("return repository save error while removing source", func(t *testing.T) {
		var (
			expectedErr      = errors.New("save failed")
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &sourceRepository{
				loadResult: source.NewSources(configuredSource),
				saveErr:    expectedErr,
			}
			sut = newSut(t, app.WithSourceRepository(repository))
		)

		_, _, err := sut.RemoveSource(configuredSource.Locator())

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestRefreshSources(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newSut = func(t *testing.T, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(testRuntime(t), optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("refresh each configured source mirror", func(t *testing.T) {
		var (
			ctx            = context.Background()
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &sourceRepository{loadResult: source.NewSources(reviewerSource, rootSource)}
			refresher      = &sourceRefresher{}
			sut            = newSut(t, app.WithSourceRepository(repository), app.WithSourceRefresher(refresher))
		)

		refresher.refreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			action := "cloned"
			if mirror.ID() == reviewerSource.ID() {
				action = "pulled"
			}

			return source.RefreshResult{Mirror: mirror, Action: action}, nil
		}

		got, err := sut.RefreshSources(ctx)

		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, rootSource, got[0].Mirror.Source)
		assert.Equal(t, "cloned", got[0].Action)
		assert.Equal(t, reviewerSource, got[1].Mirror.Source)
		assert.Equal(t, "pulled", got[1].Action)
		require.Len(t, refresher.calls, 2)
	})

	t.Run("continue after refresh error and return joined error", func(t *testing.T) {
		var (
			ctx            = context.Background()
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &sourceRepository{loadResult: source.NewSources(reviewerSource, rootSource)}
			refresher      = &sourceRefresher{}
			sut            = newSut(t, app.WithSourceRepository(repository), app.WithSourceRefresher(refresher))
		)

		refresher.refreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			if mirror.ID() == rootSource.ID() {
				return source.RefreshResult{}, errors.New("clone failed")
			}

			return source.RefreshResult{Mirror: mirror, Action: "pulled"}, nil
		}

		got, err := sut.RefreshSources(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), rootSource.ID())
		assert.Equal(t, []source.RefreshResult{{
			Mirror: got[0].Mirror,
			Action: "pulled",
		}}, got)
		assert.Equal(t, reviewerSource, got[0].Mirror.Source)
	})
}

func TestRebuildCatalog(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			discoveredSkill, err := catalog.NewSkill(sourceID, relativePath, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
		newSut = func(t *testing.T, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(testRuntime(t), optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("scan mirrors and persist catalog snapshot", func(t *testing.T) {
		var (
			indexedAt      = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &sourceRepository{loadResult: source.NewSources(reviewerSource, rootSource)}
			catalogRepo    = &catalogRepository{}
			sut            = newSut(
				t,
				app.WithSourceRepository(repository),
				app.WithCatalogRepository(catalogRepo),
				app.WithClock(clock{now: indexedAt}),
				app.WithCatalogScanner(func(mirror source.Mirror) (catalog.Skills, error) {
					switch mirror.ID() {
					case rootSource.ID():
						return catalog.NewSkills(newSkill(t, mirror.ID(), "programmer", "Programmer")), nil
					case reviewerSource.ID():
						return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
					default:
						return nil, nil
					}
				}),
			)
		)

		got, err := sut.RebuildCatalog()

		require.NoError(t, err)
		assert.Equal(t, indexedAt.UTC(), got.IndexedAt())
		require.Len(t, catalogRepo.saveCalls, 1)
		assert.Equal(t, got.IndexedAt(), catalogRepo.saveCalls[0].IndexedAt())
		assert.Equal(t, got.Skills(), catalogRepo.saveCalls[0].Skills())
		require.Len(t, got.Skills(), 2)
	})

	t.Run("save partial catalog and return joined error when one scan fails", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &sourceRepository{loadResult: source.NewSources(reviewerSource, rootSource)}
			catalogRepo    = &catalogRepository{}
			sut            = newSut(
				t,
				app.WithSourceRepository(repository),
				app.WithCatalogRepository(catalogRepo),
				app.WithClock(clock{now: time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)}),
				app.WithCatalogScanner(func(mirror source.Mirror) (catalog.Skills, error) {
					if mirror.ID() == rootSource.ID() {
						return nil, errors.New("scan failed")
					}

					return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
				}),
			)
		)

		got, err := sut.RebuildCatalog()

		require.Error(t, err)
		assert.Contains(t, err.Error(), rootSource.ID())
		require.Len(t, catalogRepo.saveCalls, 1)
		require.Len(t, got.Skills(), 1)
		assert.Equal(t, reviewerSource.ID(), got.Skills()[0].SourceID())
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
