package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncTarget(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
		newTarget = func(t *testing.T, adapter string, rootPath string) skillsync.Target {
			target, err := skillsync.NewTarget(adapter, rootPath, func(identity skill_identity.Identity) string {
				return filepath.Join(rootPath, filepath.FromSlash(identity.RelativePath()))
			})
			require.NoError(t, err)
			return target
		}
		newManifest = func(t *testing.T, adapter string, rootPath string, identities ...skill_identity.Identity) skillsync.Manifest {
			manifest, err := skillsync.NewManifest(adapter, rootPath, identities...)
			require.NoError(t, err)
			return manifest
		}
		newResolver = func(paths map[string]string) skillsync.Resolver {
			return func(identity skill_identity.Identity) (string, error) {
				path, ok := paths[identity.Key()]
				if !ok {
					return "", os.ErrNotExist
				}

				return path, nil
			}
		}
	)

	t.Run("link desired identities and unlink deselected identities", func(t *testing.T) {
		var (
			root            = t.TempDir()
			sourceRoot      = filepath.Join(root, "sources")
			targetRoot      = filepath.Join(root, "opencode")
			desiredIdentity = newIdentity(t, "source", "reviewer")
			staleIdentity   = newIdentity(t, "source", "old-reviewer")
			desiredSource   = filepath.Join(sourceRoot, "reviewer")
			staleTarget     = filepath.Join(targetRoot, "old-reviewer")
			expectedTarget  = filepath.Join(targetRoot, "reviewer")
		)

		require.NoError(t, os.MkdirAll(desiredSource, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot, 0o755))
		require.NoError(t, os.Symlink(filepath.Join(sourceRoot, "gone"), staleTarget))

		result, manifest, err := skillsync.SyncTarget(
			skill_identity.Identities{desiredIdentity},
			newTarget(t, "opencode", targetRoot),
			newManifest(t, "opencode", targetRoot, staleIdentity),
			newResolver(map[string]string{desiredIdentity.Key(): desiredSource}),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Linked)
		assert.Equal(t, 1, result.Removed)
		assert.Equal(t, skill_identity.Identities{desiredIdentity}, manifest.Identities())

		linkTarget, err := os.Readlink(expectedTarget)
		require.NoError(t, err)
		assert.Equal(t, desiredSource, linkTarget)
		_, err = os.Lstat(staleTarget)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("skip missing source identities and keep manifest clean", func(t *testing.T) {
		var (
			root            = t.TempDir()
			targetRoot      = filepath.Join(root, "claude")
			missingIdentity = newIdentity(t, "source", "missing")
		)

		require.NoError(t, os.MkdirAll(targetRoot, 0o755))

		result, manifest, err := skillsync.SyncTarget(
			skill_identity.Identities{missingIdentity},
			newTarget(t, "claude", targetRoot),
			newManifest(t, "claude", targetRoot),
			newResolver(nil),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Skipped)
		assert.Nil(t, manifest.Identities())
	})
}

func TestRun(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
		newTarget = func(t *testing.T, adapter string, rootPath string) skillsync.Target {
			target, err := skillsync.NewTarget(adapter, rootPath, func(identity skill_identity.Identity) string {
				return filepath.Join(rootPath, filepath.FromSlash(identity.RelativePath()))
			})
			require.NoError(t, err)
			return target
		}
	)

	t.Run("sync all targets and return manifests", func(t *testing.T) {
		var (
			root            = t.TempDir()
			sourceRoot      = filepath.Join(root, "sources")
			desiredIdentity = newIdentity(t, "source", "reviewer")
			sourcePath      = filepath.Join(sourceRoot, "reviewer")
			targetRoot1     = filepath.Join(root, "opencode")
			targetRoot2     = filepath.Join(root, "claude")
		)

		require.NoError(t, os.MkdirAll(sourcePath, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot1, 0o755))
		require.NoError(t, os.MkdirAll(targetRoot2, 0o755))

		result, err := skillsync.Run(
			skill_identity.Identities{desiredIdentity},
			[]skillsync.Target{
				newTarget(t, "claude", targetRoot2),
				newTarget(t, "opencode", targetRoot1),
			},
			nil,
			func(identity skill_identity.Identity) (string, error) {
				if identity.Key() != desiredIdentity.Key() {
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
			root            = t.TempDir()
			sourceRoot      = filepath.Join(root, "sources")
			desiredIdentity = newIdentity(t, "source", "reviewer")
			sourcePath      = filepath.Join(sourceRoot, "reviewer")
			sharedRoot      = filepath.Join(root, "agents", "skills")
		)

		require.NoError(t, os.MkdirAll(sourcePath, 0o755))
		require.NoError(t, os.MkdirAll(sharedRoot, 0o755))

		result, err := skillsync.Run(
			skill_identity.Identities{desiredIdentity},
			[]skillsync.Target{
				newTarget(t, "ampcode", sharedRoot),
				newTarget(t, "codex", sharedRoot),
				newTarget(t, "cursor", sharedRoot),
			},
			nil,
			func(identity skill_identity.Identity) (string, error) {
				if identity.Key() != desiredIdentity.Key() {
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
