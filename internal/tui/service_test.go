package tui_test

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
	tui "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	type dependencies struct {
		Application *ApplicationMock
	}

	var (
		newDefaultDependencies = func() *dependencies {
			return &dependencies{
				Application: &ApplicationMock{
					ListSourcesFunc: func() (source.Sources, error) { return nil, nil },
					AddSourceFunc: func(string) (source.Sources, source.Source, error) {
						return nil, source.Source{}, nil
					},
					RemoveSourceFunc: func(string) (source.Sources, source.Source, error) {
						return nil, source.Source{}, nil
					},
					RefreshCatalogFunc: func(context.Context) (app.RefreshCatalogResult, error) {
						return app.RefreshCatalogResult{}, nil
					},
					ListCatalogFunc: func() (catalog.Catalog, error) { return catalog.Catalog{}, nil },
					SyncSkillIdentitiesFunc: func(skillidentity.Identities) (skillsync.Result, error) {
						return skillsync.Result{}, nil
					},
					ListSyncManifestsFunc: func() ([]skillsync.Manifest, error) { return nil, nil },
				},
			}
		}
		newSut = func(t *testing.T, deps *dependencies) *tui.Service {
			t.Helper()

			sut, err := tui.NewService(testRuntime(t), deps.Application)
			require.NoError(t, err)

			return sut
		}
	)

	t.Run("load projects active selection from manifests with warnings", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			activeIdentity   = newIdentity(t, configuredSource.ID(), "reviewer")
			staleIdentity    = newIdentity(t, "missing-source", "writer")
			deps             = newDefaultDependencies()
			sut              = newSut(t, deps)
		)

		deps.Application.ListSourcesFunc = func() (source.Sources, error) {
			return source.Sources{configuredSource}, nil
		}
		deps.Application.ListCatalogFunc = func() (catalog.Catalog, error) {
			return catalog.NewCatalog(time.Now(), newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")), nil
		}
		deps.Application.ListSyncManifestsFunc = func() ([]skillsync.Manifest, error) {
			return []skillsync.Manifest{
				newManifest(t, "ampcode", "/tmp/agents", activeIdentity),
				newManifest(t, "opencode", "/tmp/opencode", staleIdentity),
			}, nil
		}

		snapshot, err := sut.Load(newCtx())

		require.NoError(t, err)
		assert.Equal(t, skillidentity.NewIdentities(activeIdentity, staleIdentity), snapshot.ActiveSelection)
		require.Len(t, snapshot.Warnings, 2)
		assert.Contains(t, snapshot.Warnings[0], "disagree")
		assert.Contains(t, snapshot.Warnings[1], "removed sources")
		require.Len(t, deps.Application.ListSourcesCalls(), 1)
		require.Len(t, deps.Application.ListCatalogCalls(), 1)
		require.Len(t, deps.Application.ListSyncManifestsCalls(), 1)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
	})

	t.Run("add source refreshes and reloads snapshot", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			deps             = newDefaultDependencies()
			sut              = newSut(t, deps)
		)

		deps.Application.AddSourceFunc = func(gotLocator string) (source.Sources, source.Source, error) {
			return nil, configuredSource, nil
		}
		deps.Application.RefreshCatalogFunc = func(context.Context) (app.RefreshCatalogResult, error) {
			return app.RefreshCatalogResult{}, errors.New("refresh failed")
		}
		deps.Application.ListSourcesFunc = func() (source.Sources, error) {
			return source.Sources{configuredSource}, nil
		}
		deps.Application.ListCatalogFunc = func() (catalog.Catalog, error) {
			return catalog.NewCatalog(time.Now(), newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")), nil
		}

		result, err := sut.AddSource(newCtx(), locator)

		require.Error(t, err)
		require.NotNil(t, result.Snapshot)
		assert.Contains(t, result.Summary, "Added")
		assert.Contains(t, result.Summary, locator)
		assert.Contains(t, result.Summary, "indexed 1 skill")
		require.Len(t, deps.Application.AddSourceCalls(), 1)
		assert.Equal(t, locator, deps.Application.AddSourceCalls()[0].S)
		require.Len(t, deps.Application.RefreshCatalogCalls(), 1)
		require.Len(t, deps.Application.ListSourcesCalls(), 1)
		require.Len(t, deps.Application.ListCatalogCalls(), 1)
		require.Len(t, deps.Application.ListSyncManifestsCalls(), 1)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
	})

	t.Run("sync reloads persisted manifests after the sync attempt", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			identity         = newIdentity(t, configuredSource.ID(), "reviewer")
			manifest         = newManifest(t, "opencode", "/tmp/opencode", identity)
			deps             = newDefaultDependencies()
			sut              = newSut(t, deps)
		)

		deps.Application.SyncSkillIdentitiesFunc = func(desired skillidentity.Identities) (skillsync.Result, error) {
			return skillsync.Result{DesiredCount: len(desired)}, errors.New("boom")
		}
		deps.Application.ListSourcesFunc = func() (source.Sources, error) {
			return source.Sources{configuredSource}, nil
		}
		deps.Application.ListSyncManifestsFunc = func() ([]skillsync.Manifest, error) {
			return []skillsync.Manifest{manifest}, nil
		}

		result, err := sut.Sync(newCtx(), skillidentity.NewIdentities(identity))

		require.Error(t, err)
		assert.True(t, result.HasState)
		assert.Equal(t, skillidentity.NewIdentities(identity), result.ActiveSelection)
		assert.Equal(t, []skillsync.Manifest{manifest}, result.Manifests)
		require.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 1)
		assert.Equal(t, skillidentity.NewIdentities(identity), deps.Application.SyncSkillIdentitiesCalls()[0].Identities)
		require.Len(t, deps.Application.ListSourcesCalls(), 1)
		require.Len(t, deps.Application.ListSyncManifestsCalls(), 1)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
	})
}

func newCtx() context.Context {
	return context.Background()
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

func parseSource(t *testing.T, locator string) source.Source {
	t.Helper()

	configuredSource, err := source.Parse(locator)
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
	discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
	require.NoError(t, err)

	return discoveredSkill
}

func newManifest(t *testing.T, adapter string, rootPath string, identities ...skillidentity.Identity) skillsync.Manifest {
	t.Helper()

	manifest, err := skillsync.NewManifest(adapter, rootPath, identities...)
	require.NoError(t, err)

	return manifest
}
