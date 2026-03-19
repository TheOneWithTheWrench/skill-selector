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
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create app with default dependencies", func(t *testing.T) {
		sut, err := app.New(testRuntime(t))

		require.NoError(t, err)
		require.NotNil(t, sut)
	})

	t.Run("create app with injected dependencies", func(t *testing.T) {
		sut, err := app.New(
			testRuntime(t),
			app.WithSourceRepository(&SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return nil, nil },
				SaveFunc: func(source.Sources) error { return nil },
			}),
			app.WithSourceRefresher(&SourceRefresherMock{
				RefreshFunc: func(context.Context, source.Mirror) (source.RefreshResult, error) { return source.RefreshResult{}, nil },
			}),
			app.WithCatalogRepository(&CatalogRepositoryMock{
				LoadFunc: func() (catalog.Catalog, error) { return catalog.Catalog{}, nil },
				SaveFunc: func(catalog.Catalog) error { return nil },
			}),
			app.WithCatalogScanner(func(mirror source.Mirror) (catalog.Skills, error) {
				return nil, nil
			}),
			app.WithSyncManifestRepository(&SyncManifestRepositoryMock{
				LoadAllFunc: func() ([]skillsync.Manifest, error) { return nil, nil },
				SaveFunc:    func(skillsync.Manifest) error { return nil },
			}),
			app.WithSyncTargetsLoader(func() ([]skillsync.Target, error) {
				return nil, nil
			}),
			app.WithClock(&ClockMock{NowFunc: func() time.Time {
				return time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			}}),
		)

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
			currentSources   = source.NewSources(configuredSource)
			repository       = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return currentSources, nil },
				SaveFunc: func(configuredSources source.Sources) error { currentSources = configuredSources; return nil },
			}
			sut = newSut(t, app.WithSourceRepository(repository))
		)

		got, err := sut.ListSources()

		require.NoError(t, err)
		assert.Equal(t, source.Sources{configuredSource}, got)
	})

	t.Run("add and list sources", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			currentSources   source.Sources
			repository       = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return currentSources, nil },
				SaveFunc: func(configuredSources source.Sources) error { currentSources = configuredSources; return nil },
			}
			sut = newSut(t, app.WithSourceRepository(repository))
		)

		configuredSources, addedSource, err := sut.AddSource(locator)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, addedSource)
		assert.Equal(t, source.Sources{configuredSource}, configuredSources)
		require.Len(t, repository.SaveCalls(), 1)
		assert.Equal(t, source.Sources{configuredSource}, repository.SaveCalls()[0].Sources)

		loadedSources, err := sut.ListSources()
		require.NoError(t, err)
		assert.Equal(t, configuredSources, loadedSources)
	})

	t.Run("remove source", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			currentSources   = source.NewSources(configuredSource)
			repository       = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return currentSources, nil },
				SaveFunc: func(configuredSources source.Sources) error { currentSources = configuredSources; return nil },
			}
			sut = newSut(t, app.WithSourceRepository(repository))
		)

		configuredSources, removedSource, err := sut.RemoveSource(locator)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, removedSource)
		assert.Nil(t, configuredSources)
		require.Len(t, repository.SaveCalls(), 1)
		assert.Nil(t, repository.SaveCalls()[0].Sources)
	})

	t.Run("return repository load error while adding source", func(t *testing.T) {
		var (
			expectedErr = errors.New("load failed")
			repository  = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return nil, expectedErr },
				SaveFunc: func(source.Sources) error { return nil },
			}
			sut = newSut(t, app.WithSourceRepository(repository))
		)

		_, _, err := sut.AddSource("https://github.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("return repository save error while removing source", func(t *testing.T) {
		var (
			expectedErr      = errors.New("save failed")
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(configuredSource), nil },
				SaveFunc: func(source.Sources) error { return expectedErr },
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
			repository     = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(reviewerSource, rootSource), nil },
				SaveFunc: func(source.Sources) error { return nil },
			}
			refresher = &SourceRefresherMock{}
			sut       = newSut(t, app.WithSourceRepository(repository), app.WithSourceRefresher(refresher))
		)

		refresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
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
		require.Len(t, refresher.RefreshCalls(), 2)
	})

	t.Run("continue after refresh error and return joined error", func(t *testing.T) {
		var (
			ctx            = context.Background()
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(reviewerSource, rootSource), nil },
				SaveFunc: func(source.Sources) error { return nil },
			}
			refresher = &SourceRefresherMock{}
			sut       = newSut(t, app.WithSourceRepository(repository), app.WithSourceRefresher(refresher))
		)

		refresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
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
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
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
			repository     = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(reviewerSource, rootSource), nil },
				SaveFunc: func(source.Sources) error { return nil },
			}
			persistedCatalog catalog.Catalog
			catalogRepo      = &CatalogRepositoryMock{
				LoadFunc: func() (catalog.Catalog, error) { return persistedCatalog, nil },
				SaveFunc: func(current catalog.Catalog) error { persistedCatalog = current; return nil },
			}
			sut = newSut(
				t,
				app.WithSourceRepository(repository),
				app.WithCatalogRepository(catalogRepo),
				app.WithClock(&ClockMock{NowFunc: func() time.Time { return indexedAt }}),
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
		require.Len(t, catalogRepo.SaveCalls(), 1)
		assert.Equal(t, got.IndexedAt(), catalogRepo.SaveCalls()[0].CatalogMoqParam.IndexedAt())
		assert.Equal(t, got.Skills(), catalogRepo.SaveCalls()[0].CatalogMoqParam.Skills())
		require.Len(t, got.Skills(), 2)
	})

	t.Run("save partial catalog and return joined error when one scan fails", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			repository     = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(reviewerSource, rootSource), nil },
				SaveFunc: func(source.Sources) error { return nil },
			}
			persistedCatalog catalog.Catalog
			catalogRepo      = &CatalogRepositoryMock{
				LoadFunc: func() (catalog.Catalog, error) { return persistedCatalog, nil },
				SaveFunc: func(current catalog.Catalog) error { persistedCatalog = current; return nil },
			}
			sut = newSut(
				t,
				app.WithSourceRepository(repository),
				app.WithCatalogRepository(catalogRepo),
				app.WithClock(&ClockMock{NowFunc: func() time.Time { return time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC) }}),
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
		require.Len(t, catalogRepo.SaveCalls(), 1)
		require.Len(t, got.Skills(), 1)
		assert.Equal(t, reviewerSource.ID(), got.Skills()[0].SourceID())
	})
}

func TestCatalog(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
		newSut = func(t *testing.T, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(testRuntime(t), optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("list persisted catalog", func(t *testing.T) {
		var (
			reviewerSkill = newSkill(t, "source-a", "reviewer", "Reviewer")
			repository    = &CatalogRepositoryMock{
				LoadFunc: func() (catalog.Catalog, error) { return catalog.NewCatalog(time.Time{}, reviewerSkill), nil },
				SaveFunc: func(catalog.Catalog) error { return nil },
			}
			sut = newSut(t, app.WithCatalogRepository(repository))
		)

		got, err := sut.ListCatalog()

		require.NoError(t, err)
		assert.Equal(t, catalog.Skills{reviewerSkill}, got.Skills())
	})

	t.Run("refresh catalog uses refreshed mirrors and persists rebuilt snapshot", func(t *testing.T) {
		var (
			ctx              = context.Background()
			indexedAt        = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &SourceRepositoryMock{
				LoadFunc: func() (source.Sources, error) { return source.NewSources(configuredSource), nil },
				SaveFunc: func(source.Sources) error { return nil },
			}
			refresher        = &SourceRefresherMock{}
			persistedCatalog catalog.Catalog
			catalogRepo      = &CatalogRepositoryMock{
				LoadFunc: func() (catalog.Catalog, error) { return persistedCatalog, nil },
				SaveFunc: func(current catalog.Catalog) error { persistedCatalog = current; return nil },
			}
			sut = newSut(
				t,
				app.WithSourceRepository(repository),
				app.WithSourceRefresher(refresher),
				app.WithCatalogRepository(catalogRepo),
				app.WithClock(&ClockMock{NowFunc: func() time.Time { return indexedAt }}),
				app.WithCatalogScanner(func(mirror source.Mirror) (catalog.Skills, error) {
					return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
				}),
			)
		)

		refresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			return source.RefreshResult{Mirror: mirror, Action: "cloned"}, nil
		}

		got, err := sut.RefreshCatalog(ctx)

		require.NoError(t, err)
		require.Len(t, got.Sources, 1)
		assert.Equal(t, configuredSource, got.Sources[0].Mirror.Source)
		assert.Equal(t, "cloned", got.Sources[0].Action)
		assert.Equal(t, indexedAt, got.Catalog.IndexedAt())
		require.Len(t, catalogRepo.SaveCalls(), 1)
		assert.Equal(t, got.Catalog.Skills(), catalogRepo.SaveCalls()[0].CatalogMoqParam.Skills())
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
