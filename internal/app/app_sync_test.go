package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type syncManifestRepository struct {
	loadResult []skillsync.Manifest
	loadErr    error
	saveErr    error
	saveCalls  []skillsync.Manifest
}

func (r *syncManifestRepository) LoadAll() ([]skillsync.Manifest, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}

	return append([]skillsync.Manifest(nil), r.loadResult...), nil
}

func (r *syncManifestRepository) Save(manifest skillsync.Manifest) error {
	r.saveCalls = append(r.saveCalls, manifest)
	if r.saveErr != nil {
		return r.saveErr
	}

	r.loadResult = append([]skillsync.Manifest(nil), r.saveCalls...)
	return nil
}

func TestListSyncManifests(t *testing.T) {
	var (
		newSut = func(t *testing.T, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(testRuntime(t), optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("list persisted sync manifests", func(t *testing.T) {
		manifest, err := skillsync.NewManifest("opencode", "/tmp/opencode")
		require.NoError(t, err)

		repository := &syncManifestRepository{loadResult: []skillsync.Manifest{manifest}}
		sut := newSut(t, app.WithSyncManifestRepository(repository))

		got, err := sut.ListSyncManifests()

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, manifest.Adapter(), got[0].Adapter())
		assert.Equal(t, manifest.RootPath(), got[0].RootPath())
	})
}

func TestSyncSkillIdentities(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
		newTarget = func(t *testing.T, adapter string, rootPath string) skillsync.Target {
			target, err := skillsync.NewTarget(adapter, rootPath, func(identity skillidentity.Identity) string {
				return filepath.Join(rootPath, filepath.FromSlash(identity.RelativePath()))
			})
			require.NoError(t, err)
			return target
		}
		newSut = func(t *testing.T, runtime paths.Runtime, optionFuncs ...app.Option) *app.App {
			sut, err := app.New(runtime, optionFuncs...)
			require.NoError(t, err)
			return sut
		}
	)

	t.Run("sync identities across detected targets and save manifests", func(t *testing.T) {
		var (
			runtime          = testRuntime(t)
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &sourceRepository{loadResult: source.NewSources(configuredSource)}
			manifestRepo     = &syncManifestRepository{}
			targetRoot       = filepath.Join(t.TempDir(), "opencode")
			target           = newTarget(t, "opencode", targetRoot)
			identity         = newIdentity(t, configuredSource.ID(), "reviewer")
		)

		mirror, err := source.NewMirror(configuredSource, runtime.SourcesDir)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(mirror.SkillPath(identity.RelativePath()), 0o755))

		sut := newSut(
			t,
			runtime,
			app.WithSourceRepository(repository),
			app.WithSyncManifestRepository(manifestRepo),
			app.WithSyncTargetsLoader(func() ([]skillsync.Target, error) {
				return []skillsync.Target{target}, nil
			}),
		)

		got, err := sut.SyncSkillIdentities(skillidentity.NewIdentities(identity))

		require.NoError(t, err)
		require.Len(t, got.Targets, 1)
		assert.Equal(t, 1, got.Targets[0].Linked)
		require.Len(t, got.Manifests, 1)
		require.Len(t, manifestRepo.saveCalls, 1)
		assert.Equal(t, skillidentity.NewIdentities(identity), manifestRepo.saveCalls[0].Identities())

		linkTarget, err := os.Readlink(filepath.Join(targetRoot, "reviewer"))
		require.NoError(t, err)
		assert.Equal(t, mirror.SkillPath(identity.RelativePath()), linkTarget)
	})

	t.Run("return target loader error", func(t *testing.T) {
		var (
			expectedErr  = errors.New("target load failed")
			runtime      = testRuntime(t)
			manifestRepo = &syncManifestRepository{}
			sut          = newSut(
				t,
				runtime,
				app.WithSyncManifestRepository(manifestRepo),
				app.WithSyncTargetsLoader(func() ([]skillsync.Target, error) {
					return nil, expectedErr
				}),
			)
		)

		_, err := sut.SyncSkillIdentities(nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("return save error after sync succeeds", func(t *testing.T) {
		var (
			runtime          = testRuntime(t)
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			repository       = &sourceRepository{loadResult: source.NewSources(configuredSource)}
			expectedErr      = errors.New("save failed")
			manifestRepo     = &syncManifestRepository{saveErr: expectedErr}
			targetRoot       = filepath.Join(t.TempDir(), "opencode")
			target           = newTarget(t, "opencode", targetRoot)
			identity         = newIdentity(t, configuredSource.ID(), "reviewer")
		)

		mirror, err := source.NewMirror(configuredSource, runtime.SourcesDir)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(mirror.SkillPath(identity.RelativePath()), 0o755))

		sut := newSut(
			t,
			runtime,
			app.WithSourceRepository(repository),
			app.WithSyncManifestRepository(manifestRepo),
			app.WithSyncTargetsLoader(func() ([]skillsync.Target, error) {
				return []skillsync.Target{target}, nil
			}),
		)

		got, err := sut.SyncSkillIdentities(skillidentity.NewIdentities(identity))

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		require.Len(t, got.Manifests, 1)
		require.Len(t, manifestRepo.saveCalls, 1)
	})
}
