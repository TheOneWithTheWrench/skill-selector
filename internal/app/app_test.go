package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
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
		var (
			deps = newDefaultDependencies()
			sut  = newSut(t, deps)
		)

		require.NotNil(t, sut)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}

func TestSources(t *testing.T) {
	t.Run("list sources", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			deps             = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(configuredSource), nil
		}

		sut := newSut(t, deps)

		got, err := sut.ListSources()

		require.NoError(t, err)
		assert.Equal(t, source.Sources{configuredSource}, got)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 1)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("add source", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			deps             = newDefaultDependencies()
			storedSources    source.Sources
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return storedSources, nil
		}
		deps.SourceRepository.SaveFunc = func(configuredSources source.Sources) error {
			storedSources = configuredSources
			return nil
		}

		sut := newSut(t, deps)

		configuredSources, addedSource, err := sut.AddSource(locator)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, addedSource)
		assert.Equal(t, source.Sources{configuredSource}, configuredSources)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.SourceRepository.SaveCalls(), 1)
		assert.Equal(t, source.Sources{configuredSource}, deps.SourceRepository.SaveCalls()[0].Sources)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("remove source", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			deps             = newDefaultDependencies()
			storedSources    = source.NewSources(configuredSource)
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return storedSources, nil
		}
		deps.SourceRepository.SaveFunc = func(configuredSources source.Sources) error {
			storedSources = configuredSources
			return nil
		}

		sut := newSut(t, deps)

		configuredSources, removedSource, err := sut.RemoveSource(locator)

		require.NoError(t, err)
		assert.Equal(t, configuredSource, removedSource)
		assert.Nil(t, configuredSources)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.SourceRepository.SaveCalls(), 1)
		assert.Nil(t, deps.SourceRepository.SaveCalls()[0].Sources)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("return repository load error while adding source", func(t *testing.T) {
		var (
			expectedErr = errors.New("load failed")
			deps        = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return nil, expectedErr
		}

		sut := newSut(t, deps)

		_, _, err := sut.AddSource("https://github.com/anthropics/skills/tree/main/skills")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 1)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("return repository save error while removing source", func(t *testing.T) {
		var (
			expectedErr      = errors.New("save failed")
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			deps             = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(configuredSource), nil
		}
		deps.SourceRepository.SaveFunc = func(source.Sources) error {
			return expectedErr
		}

		sut := newSut(t, deps)

		_, _, err := sut.RemoveSource(configuredSource.Locator())

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 1)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}

func TestRefreshSources(t *testing.T) {
	t.Run("refresh each configured source mirror", func(t *testing.T) {
		var (
			ctx            = newCtx()
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			deps           = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(reviewerSource, rootSource), nil
		}
		deps.SourceRefresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			action := "cloned"
			if mirror.ID() == reviewerSource.ID() {
				action = "pulled"
			}

			return source.RefreshResult{Mirror: mirror, Action: action}, nil
		}

		sut := newSut(t, deps)

		got, err := sut.RefreshSources(ctx)

		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, rootSource, got[0].Mirror.Source)
		assert.Equal(t, "cloned", got[0].Action)
		assert.Equal(t, reviewerSource, got[1].Mirror.Source)
		assert.Equal(t, "pulled", got[1].Action)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.SourceRefresher.RefreshCalls(), 2)
		assert.Equal(t, rootSource.ID(), deps.SourceRefresher.RefreshCalls()[0].Mirror.ID())
		assert.Equal(t, reviewerSource.ID(), deps.SourceRefresher.RefreshCalls()[1].Mirror.ID())
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("continue after refresh error and return joined error", func(t *testing.T) {
		var (
			ctx            = newCtx()
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			deps           = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(reviewerSource, rootSource), nil
		}
		deps.SourceRefresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			if mirror.ID() == rootSource.ID() {
				return source.RefreshResult{}, errors.New("clone failed")
			}

			return source.RefreshResult{Mirror: mirror, Action: "pulled"}, nil
		}

		sut := newSut(t, deps)

		got, err := sut.RefreshSources(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), rootSource.ID())
		assert.Equal(t, []source.RefreshResult{{
			Mirror: got[0].Mirror,
			Action: "pulled",
		}}, got)
		assert.Equal(t, reviewerSource, got[0].Mirror.Source)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.SourceRefresher.RefreshCalls(), 2)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}

func TestRebuildCatalog(t *testing.T) {
	t.Run("scan mirrors and persist catalog snapshot", func(t *testing.T) {
		var (
			indexedAt      = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			deps           = newDefaultDependencies()
			scannerCalls   []source.Mirror
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(reviewerSource, rootSource), nil
		}
		deps.Clock.NowFunc = func() time.Time { return indexedAt }
		deps.CatalogScanner = func(mirror source.Mirror) (catalog.Skills, error) {
			scannerCalls = append(scannerCalls, mirror)

			switch mirror.ID() {
			case rootSource.ID():
				return catalog.NewSkills(newSkill(t, mirror.ID(), "programmer", "Programmer")), nil
			case reviewerSource.ID():
				return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
			default:
				return nil, nil
			}
		}

		sut := newSut(t, deps)

		got, err := sut.RebuildCatalog()

		require.NoError(t, err)
		assert.Equal(t, indexedAt.UTC(), got.IndexedAt())
		require.Len(t, scannerCalls, 2)
		assert.Equal(t, rootSource.ID(), scannerCalls[0].ID())
		assert.Equal(t, reviewerSource.ID(), scannerCalls[1].ID())
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.CatalogRepository.SaveCalls(), 1)
		assert.Equal(t, got.IndexedAt(), deps.CatalogRepository.SaveCalls()[0].CatalogMoqParam.IndexedAt())
		assert.Equal(t, got.Skills(), deps.CatalogRepository.SaveCalls()[0].CatalogMoqParam.Skills())
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		require.Len(t, deps.Clock.NowCalls(), 1)
	})

	t.Run("save partial catalog and return joined error when one scan fails", func(t *testing.T) {
		var (
			rootSource     = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			reviewerSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills/reviewer")
			deps           = newDefaultDependencies()
			scannerCalls   []source.Mirror
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(reviewerSource, rootSource), nil
		}
		deps.CatalogScanner = func(mirror source.Mirror) (catalog.Skills, error) {
			scannerCalls = append(scannerCalls, mirror)
			if mirror.ID() == rootSource.ID() {
				return nil, errors.New("scan failed")
			}

			return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
		}

		sut := newSut(t, deps)

		got, err := sut.RebuildCatalog()

		require.Error(t, err)
		assert.Contains(t, err.Error(), rootSource.ID())
		require.Len(t, scannerCalls, 2)
		require.Len(t, got.Skills(), 1)
		assert.Equal(t, reviewerSource.ID(), got.Skills()[0].SourceID())
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.CatalogRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		require.Len(t, deps.Clock.NowCalls(), 1)
	})
}

func TestCatalog(t *testing.T) {
	t.Run("list persisted catalog", func(t *testing.T) {
		var (
			reviewerSkill = newSkill(t, "source-a", "reviewer", "Reviewer")
			deps          = newDefaultDependencies()
		)

		deps.CatalogRepository.LoadFunc = func() (catalog.Catalog, error) {
			return catalog.NewCatalog(time.Time{}, reviewerSkill), nil
		}

		sut := newSut(t, deps)

		got, err := sut.ListCatalog()

		require.NoError(t, err)
		assert.Equal(t, catalog.Skills{reviewerSkill}, got.Skills())
		require.Len(t, deps.CatalogRepository.LoadCalls(), 1)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("refresh catalog uses refreshed mirrors and persists rebuilt snapshot", func(t *testing.T) {
		var (
			ctx              = newCtx()
			indexedAt        = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			deps             = newDefaultDependencies()
			scannerCalls     []source.Mirror
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(configuredSource), nil
		}
		deps.SourceRefresher.RefreshFunc = func(ctx context.Context, mirror source.Mirror) (source.RefreshResult, error) {
			return source.RefreshResult{Mirror: mirror, Action: "cloned"}, nil
		}
		deps.Clock.NowFunc = func() time.Time { return indexedAt }
		deps.CatalogScanner = func(mirror source.Mirror) (catalog.Skills, error) {
			scannerCalls = append(scannerCalls, mirror)
			return catalog.NewSkills(newSkill(t, mirror.ID(), "reviewer", "Reviewer")), nil
		}

		sut := newSut(t, deps)

		got, err := sut.RefreshCatalog(ctx)

		require.NoError(t, err)
		require.Len(t, got.Sources, 1)
		assert.Equal(t, configuredSource, got.Sources[0].Mirror.Source)
		assert.Equal(t, "cloned", got.Sources[0].Action)
		assert.Equal(t, indexedAt, got.Catalog.IndexedAt())
		require.Len(t, scannerCalls, 1)
		assert.Equal(t, configuredSource.ID(), scannerCalls[0].ID())
		require.Len(t, deps.SourceRepository.LoadCalls(), 2)
		require.Len(t, deps.SourceRefresher.RefreshCalls(), 1)
		require.Len(t, deps.CatalogRepository.SaveCalls(), 1)
		assert.Equal(t, got.Catalog.Skills(), deps.CatalogRepository.SaveCalls()[0].CatalogMoqParam.Skills())
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		require.Len(t, deps.Clock.NowCalls(), 1)
	})
}
