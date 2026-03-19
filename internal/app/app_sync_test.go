package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSyncManifests(t *testing.T) {
	t.Run("list persisted sync manifests", func(t *testing.T) {
		var (
			manifest, err = sync.NewManifest("opencode", "/tmp/opencode")
			deps          = newDefaultDependencies()
		)
		require.NoError(t, err)

		deps.SyncManifestRepo.LoadAllFunc = func() ([]sync.Manifest, error) {
			return []sync.Manifest{manifest}, nil
		}

		sut := newSut(t, deps)

		got, err := sut.ListSyncManifests()

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, manifest.Adapter(), got[0].Adapter())
		assert.Equal(t, manifest.RootPath(), got[0].RootPath())
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 1)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}

func TestSyncSkillIdentities(t *testing.T) {
	t.Run("sync identities across detected targets and save manifests", func(t *testing.T) {
		var (
			runtime            = testRuntime(t)
			configuredSource   = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			identity           = newIdentity(t, configuredSource.ID(), "reviewer")
			persistedManifests []sync.Manifest
			syncTargetsCalls   int
			deps               = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(configuredSource), nil
		}
		deps.SyncManifestRepo.LoadAllFunc = func() ([]sync.Manifest, error) {
			return append([]sync.Manifest(nil), persistedManifests...), nil
		}
		deps.SyncManifestRepo.SaveFunc = func(manifest sync.Manifest) error {
			persistedManifests = append(persistedManifests, manifest)
			return nil
		}

		targetRoot := filepath.Join(t.TempDir(), "opencode")
		target := newTarget(t, "opencode", targetRoot)
		deps.SyncTargetsLoader = func() ([]sync.Target, error) {
			syncTargetsCalls++
			return []sync.Target{target}, nil
		}

		mirror, err := source.NewMirror(configuredSource, runtime.SourcesDir)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(mirror.SkillPath(identity.RelativePath()), 0o755))

		sut := newSutWithRuntime(t, deps, runtime)

		got, err := sut.SyncSkillIdentities(skillidentity.NewIdentities(identity))

		require.NoError(t, err)
		require.Len(t, got.Targets, 1)
		assert.Equal(t, 1, got.Targets[0].Linked)
		require.Len(t, got.Manifests, 1)
		require.Len(t, deps.SyncManifestRepo.SaveCalls(), 1)
		assert.Equal(t, skillidentity.NewIdentities(identity), deps.SyncManifestRepo.SaveCalls()[0].Manifest.Identities())
		assert.Equal(t, 1, syncTargetsCalls)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		require.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 1)
		assert.Len(t, deps.Clock.NowCalls(), 0)

		linkTarget, err := os.Readlink(filepath.Join(targetRoot, "reviewer"))
		require.NoError(t, err)
		assert.Equal(t, mirror.SkillPath(identity.RelativePath()), linkTarget)
	})

	t.Run("return target loader error", func(t *testing.T) {
		var (
			expectedErr      = errors.New("target load failed")
			runtime          = testRuntime(t)
			syncTargetsCalls int
			deps             = newDefaultDependencies()
		)

		deps.SyncTargetsLoader = func() ([]sync.Target, error) {
			syncTargetsCalls++
			return nil, expectedErr
		}

		sut := newSutWithRuntime(t, deps, runtime)

		_, err := sut.SyncSkillIdentities(nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 1, syncTargetsCalls)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("return save error after sync succeeds", func(t *testing.T) {
		var (
			expectedErr      = errors.New("save failed")
			runtime          = testRuntime(t)
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			identity         = newIdentity(t, configuredSource.ID(), "reviewer")
			syncTargetsCalls int
			deps             = newDefaultDependencies()
		)

		deps.SourceRepository.LoadFunc = func() (source.Sources, error) {
			return source.NewSources(configuredSource), nil
		}
		deps.SyncManifestRepo.LoadAllFunc = func() ([]sync.Manifest, error) { return nil, nil }
		deps.SyncManifestRepo.SaveFunc = func(sync.Manifest) error { return expectedErr }

		targetRoot := filepath.Join(t.TempDir(), "opencode")
		target := newTarget(t, "opencode", targetRoot)
		deps.SyncTargetsLoader = func() ([]sync.Target, error) {
			syncTargetsCalls++
			return []sync.Target{target}, nil
		}

		mirror, err := source.NewMirror(configuredSource, runtime.SourcesDir)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(mirror.SkillPath(identity.RelativePath()), 0o755))

		sut := newSutWithRuntime(t, deps, runtime)

		got, err := sut.SyncSkillIdentities(skillidentity.NewIdentities(identity))

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		require.Len(t, got.Manifests, 1)
		assert.Equal(t, 1, syncTargetsCalls)
		require.Len(t, deps.SourceRepository.LoadCalls(), 1)
		require.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 1)
		require.Len(t, deps.SyncManifestRepo.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}
