package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/profile"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/require"
)

type dependencies struct {
	SourceRepository  *SourceRepositoryMock
	SourceRefresher   *SourceRefresherMock
	CatalogRepository *CatalogRepositoryMock
	ProfileRepository *ProfileRepositoryMock
	SyncManifestRepo  *SyncManifestRepositoryMock
	Clock             *ClockMock
	CatalogScanner    app.CatalogScanner
	SyncTargetsLoader app.SyncTargetsLoader
}

func newDefaultDependencies() *dependencies {
	return &dependencies{
		SourceRepository: &SourceRepositoryMock{
			LoadFunc: func() (source.Sources, error) { return nil, nil },
			SaveFunc: func(source.Sources) error { return nil },
		},
		SourceRefresher: &SourceRefresherMock{
			RefreshFunc: func(context.Context, source.Mirror) (source.RefreshResult, error) {
				return source.RefreshResult{}, nil
			},
		},
		CatalogRepository: &CatalogRepositoryMock{
			LoadFunc: func() (catalog.Catalog, error) { return catalog.Catalog{}, nil },
			SaveFunc: func(catalog.Catalog) error { return nil },
		},
		ProfileRepository: &ProfileRepositoryMock{
			LoadFunc: func() (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
			SaveFunc: func(profile.Profiles) error { return nil },
		},
		SyncManifestRepo: &SyncManifestRepositoryMock{
			LoadAllFunc: func() ([]skillsync.Manifest, error) { return nil, nil },
			SaveFunc:    func(skillsync.Manifest) error { return nil },
		},
		Clock: &ClockMock{
			NowFunc: func() time.Time {
				return time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			},
		},
		CatalogScanner:    func(source.Mirror) (catalog.Skills, error) { return nil, nil },
		SyncTargetsLoader: func() ([]skillsync.Target, error) { return nil, nil },
	}
}

func newSut(t *testing.T, deps *dependencies) *app.App {
	t.Helper()

	return newSutWithRuntime(t, deps, testRuntime(t))
}

func newSutWithRuntime(t *testing.T, deps *dependencies, runtime paths.Runtime) *app.App {
	t.Helper()

	sut, err := app.New(
		runtime,
		app.WithSourceRepository(deps.SourceRepository),
		app.WithSourceRefresher(deps.SourceRefresher),
		app.WithCatalogRepository(deps.CatalogRepository),
		app.WithProfileRepository(deps.ProfileRepository),
		app.WithCatalogScanner(deps.CatalogScanner),
		app.WithSyncManifestRepository(deps.SyncManifestRepo),
		app.WithSyncTargetsLoader(deps.SyncTargetsLoader),
		app.WithClock(deps.Clock),
	)
	require.NoError(t, err)

	return sut
}

func newCtx() context.Context {
	return context.Background()
}

func parseSource(t *testing.T, rawURL string) source.Source {
	t.Helper()

	configuredSource, err := source.Parse(rawURL)
	require.NoError(t, err)

	return configuredSource
}

func newIdentity(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
	t.Helper()

	identity, err := skillidentity.New(sourceID, relativePath)
	require.NoError(t, err)

	return identity
}

func newSkill(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
	t.Helper()

	identity := newIdentity(t, sourceID, relativePath)
	skill, err := catalog.NewSkill(identity, name, name+" description")
	require.NoError(t, err)

	return skill
}

func newTarget(t *testing.T, adapter string, rootPath string) skillsync.Target {
	t.Helper()

	target, err := skillsync.NewTarget(adapter, rootPath, func(identity skillidentity.Identity) string {
		return filepath.Join(rootPath, filepath.FromSlash(identity.RelativePath()))
	})
	require.NoError(t, err)

	return target
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
