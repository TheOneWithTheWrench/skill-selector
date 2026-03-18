package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSymlink(t *testing.T) {
	t.Run("create new symlink", func(t *testing.T) {
		var (
			root   = t.TempDir()
			target = filepath.Join(root, "target")
			link   = filepath.Join(root, "link")
		)

		require.NoError(t, os.MkdirAll(target, 0o755))

		changed, err := ensureSymlink(target, link)

		require.NoError(t, err)
		assert.True(t, changed)
		resolved, err := os.Readlink(link)
		require.NoError(t, err)
		assert.Equal(t, target, resolved)
	})

	t.Run("no change when symlink already points to correct target", func(t *testing.T) {
		var (
			root   = t.TempDir()
			target = filepath.Join(root, "target")
			link   = filepath.Join(root, "link")
		)

		require.NoError(t, os.MkdirAll(target, 0o755))
		require.NoError(t, os.Symlink(target, link))

		changed, err := ensureSymlink(target, link)

		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("replace symlink pointing to wrong target", func(t *testing.T) {
		var (
			root      = t.TempDir()
			oldTarget = filepath.Join(root, "old")
			newTarget = filepath.Join(root, "new")
			link      = filepath.Join(root, "link")
		)

		require.NoError(t, os.MkdirAll(oldTarget, 0o755))
		require.NoError(t, os.MkdirAll(newTarget, 0o755))
		require.NoError(t, os.Symlink(oldTarget, link))

		changed, err := ensureSymlink(newTarget, link)

		require.NoError(t, err)
		assert.True(t, changed)
		resolved, err := os.Readlink(link)
		require.NoError(t, err)
		assert.Equal(t, newTarget, resolved)
	})

	t.Run("refuse to overwrite non symlink file", func(t *testing.T) {
		var (
			root    = t.TempDir()
			target  = filepath.Join(root, "target")
			regular = filepath.Join(root, "regular")
		)

		require.NoError(t, os.MkdirAll(target, 0o755))
		require.NoError(t, os.WriteFile(regular, []byte("data"), 0o644))

		_, err := ensureSymlink(target, regular)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-symlink")
	})

	t.Run("create parent directories for link path", func(t *testing.T) {
		var (
			root   = t.TempDir()
			target = filepath.Join(root, "target")
			link   = filepath.Join(root, "nested", "deep", "link")
		)

		require.NoError(t, os.MkdirAll(target, 0o755))

		changed, err := ensureSymlink(target, link)

		require.NoError(t, err)
		assert.True(t, changed)
	})
}

func TestRemoveOwnedLink(t *testing.T) {
	t.Run("remove existing symlink", func(t *testing.T) {
		var (
			root   = t.TempDir()
			target = filepath.Join(root, "target")
			link   = filepath.Join(root, "link")
		)

		require.NoError(t, os.MkdirAll(target, 0o755))
		require.NoError(t, os.Symlink(target, link))

		removed, err := removeOwnedLink(link)

		require.NoError(t, err)
		assert.True(t, removed)
		_, statErr := os.Lstat(link)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("return false for nonexistent path", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "nonexistent")

		removed, err := removeOwnedLink(link)

		require.NoError(t, err)
		assert.False(t, removed)
	})

	t.Run("refuse to remove non symlink file", func(t *testing.T) {
		regular := filepath.Join(t.TempDir(), "regular")
		require.NoError(t, os.WriteFile(regular, []byte("data"), 0o644))

		_, err := removeOwnedLink(regular)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-symlink")
		assert.FileExists(t, regular)
	})
}
