package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillref"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncTarget(t *testing.T) {
	var (
		newRef = func(t *testing.T, sourceID string, relativePath string) skillref.Ref {
			ref, err := skillref.New(sourceID, relativePath)
			require.NoError(t, err)
			return ref
		}
		newTarget = func(t *testing.T, adapter string, rootPath string) skillsync.Target {
			target, err := skillsync.NewTarget(adapter, rootPath, func(ref skillref.Ref) string {
				return filepath.Join(rootPath, filepath.FromSlash(ref.RelativePath()))
			})
			require.NoError(t, err)
			return target
		}
		newManifest = func(t *testing.T, adapter string, rootPath string, refs ...skillref.Ref) skillsync.Manifest {
			manifest, err := skillsync.NewManifest(adapter, rootPath, refs...)
			require.NoError(t, err)
			return manifest
		}
		newResolver = func(paths map[string]string) skillsync.Resolver {
			return func(ref skillref.Ref) (string, error) {
				path, ok := paths[ref.Key()]
				if !ok {
					return "", os.ErrNotExist
				}

				return path, nil
			}
		}
	)

	t.Run("link desired refs and unlink deselected refs", func(t *testing.T) {
		var (
			root           = t.TempDir()
			sourceRoot     = filepath.Join(root, "sources")
			targetRoot     = filepath.Join(root, "opencode")
			desiredRef     = newRef(t, "source", "reviewer")
			staleRef       = newRef(t, "source", "old-reviewer")
			desiredSource  = filepath.Join(sourceRoot, "reviewer")
			staleTarget    = filepath.Join(targetRoot, "old-reviewer")
			expectedTarget = filepath.Join(targetRoot, "reviewer")
		)

		require.NoError(t, os.MkdirAll(desiredSource, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot, 0o755))
		require.NoError(t, os.Symlink(filepath.Join(sourceRoot, "gone"), staleTarget))

		result, manifest, err := skillsync.SyncTarget(
			skillref.Refs{desiredRef},
			newTarget(t, "opencode", targetRoot),
			newManifest(t, "opencode", targetRoot, staleRef),
			newResolver(map[string]string{desiredRef.Key(): desiredSource}),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Linked)
		assert.Equal(t, 1, result.Removed)
		assert.Equal(t, skillref.Refs{desiredRef}, manifest.Refs())

		linkTarget, err := os.Readlink(expectedTarget)
		require.NoError(t, err)
		assert.Equal(t, desiredSource, linkTarget)
		_, err = os.Lstat(staleTarget)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("skip missing source refs and keep manifest clean", func(t *testing.T) {
		var (
			root       = t.TempDir()
			targetRoot = filepath.Join(root, "claude")
			missingRef = newRef(t, "source", "missing")
		)

		require.NoError(t, os.MkdirAll(targetRoot, 0o755))

		result, manifest, err := skillsync.SyncTarget(
			skillref.Refs{missingRef},
			newTarget(t, "claude", targetRoot),
			newManifest(t, "claude", targetRoot),
			newResolver(nil),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Skipped)
		assert.Nil(t, manifest.Refs())
	})
}

func TestRun(t *testing.T) {
	var (
		newRef = func(t *testing.T, sourceID string, relativePath string) skillref.Ref {
			ref, err := skillref.New(sourceID, relativePath)
			require.NoError(t, err)
			return ref
		}
		newTarget = func(t *testing.T, adapter string, rootPath string) skillsync.Target {
			target, err := skillsync.NewTarget(adapter, rootPath, func(ref skillref.Ref) string {
				return filepath.Join(rootPath, filepath.FromSlash(ref.RelativePath()))
			})
			require.NoError(t, err)
			return target
		}
	)

	t.Run("sync all targets and return manifests", func(t *testing.T) {
		var (
			root        = t.TempDir()
			sourceRoot  = filepath.Join(root, "sources")
			desiredRef  = newRef(t, "source", "reviewer")
			sourcePath  = filepath.Join(sourceRoot, "reviewer")
			targetRoot1 = filepath.Join(root, "opencode")
			targetRoot2 = filepath.Join(root, "claude")
		)

		require.NoError(t, os.MkdirAll(sourcePath, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot1, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot2, 0o755))

		result, err := skillsync.Run(
			skillref.Refs{desiredRef},
			[]skillsync.Target{
				newTarget(t, "claude", targetRoot2),
				newTarget(t, "opencode", targetRoot1),
			},
			nil,
			func(ref skillref.Ref) (string, error) {
				if ref.Key() != desiredRef.Key() {
					return "", os.ErrNotExist
				}

				return sourcePath, nil
			},
		)

		require.NoError(t, err)
		require.Len(t, result.Targets, 2)
		require.Len(t, result.Manifests, 2)
		assert.Equal(t, "claude", result.Targets[0].Adapter)
		assert.Equal(t, "opencode", result.Targets[1].Adapter)
		assert.Equal(t, "Synced 1 selected skill to 2 locations", result.Summary())
	})

	t.Run("deduplicate targets that share the same root path", func(t *testing.T) {
		var (
			root       = t.TempDir()
			sourceRoot = filepath.Join(root, "sources")
			desiredRef = newRef(t, "source", "reviewer")
			sourcePath = filepath.Join(sourceRoot, "reviewer")
			sharedRoot = filepath.Join(root, "agents", "skills")
		)

		require.NoError(t, os.MkdirAll(sourcePath, 0o755))
		require.NoError(t, os.MkdirAll(sharedRoot, 0o755))

		result, err := skillsync.Run(
			skillref.Refs{desiredRef},
			[]skillsync.Target{
				newTarget(t, "ampcode", sharedRoot),
				newTarget(t, "codex", sharedRoot),
				newTarget(t, "cursor", sharedRoot),
			},
			nil,
			func(ref skillref.Ref) (string, error) {
				if ref.Key() != desiredRef.Key() {
					return "", os.ErrNotExist
				}

				return sourcePath, nil
			},
		)

		require.NoError(t, err)
		require.Len(t, result.Targets, 1)
		require.Len(t, result.Manifests, 3)
		assert.Equal(t, "ampcode,codex,cursor", result.Targets[0].Adapter)
		assert.Equal(t, "Synced 1 selected skill to 1 location", result.Summary())
		assert.Equal(t, sharedRoot, result.Targets[0].RootPath)
		assert.Equal(t, []string{"ampcode", "codex", "cursor"}, result.Targets[0].Adapters)
	})
}
