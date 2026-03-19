package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectoryManifestRepository(t *testing.T) {
	t.Run("return error when directory is empty", func(t *testing.T) {
		_, err := skillsync.NewDirectoryManifestRepository("")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sync manifest directory required")
	})
}

func TestDirectoryManifestRepository(t *testing.T) {
	var (
		newRepository = func(t *testing.T) (*skillsync.DirectoryManifestRepository, string) {
			dir := filepath.Join(t.TempDir(), "sync-state")
			repository, err := skillsync.NewDirectoryManifestRepository(dir)
			require.NoError(t, err)
			return repository, dir
		}
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
		newManifest = func(t *testing.T, adapter string, rootPath string, identities ...skillidentity.Identity) skillsync.Manifest {
			manifest, err := skillsync.NewManifest(adapter, rootPath, identities...)
			require.NoError(t, err)
			return manifest
		}
	)

	t.Run("load nil for missing directory", func(t *testing.T) {
		repository, _ := newRepository(t)

		got, err := repository.LoadAll()

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("save and load manifests sorted by adapter", func(t *testing.T) {
		var (
			repository, _      = newRepository(t)
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			cursorManifest     = newManifest(t, "cursor", "/tmp/cursor", reviewerIdentity)
			ampcodeManifest    = newManifest(t, "ampcode", "/tmp/ampcode", programmerIdentity)
		)

		require.NoError(t, repository.Save(cursorManifest))
		require.NoError(t, repository.Save(ampcodeManifest))

		got, err := repository.LoadAll()

		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "ampcode", got[0].Adapter())
		assert.Equal(t, "cursor", got[1].Adapter())
		assert.Equal(t, skillidentity.Identities{programmerIdentity}, got[0].Identities())
		assert.Equal(t, skillidentity.Identities{reviewerIdentity}, got[1].Identities())
	})

	t.Run("support legacy agent field when loading manifest", func(t *testing.T) {
		var (
			repository, dir = newRepository(t)
			manifestPath    = filepath.Join(dir, "opencode.json")
		)

		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(manifestPath, []byte("{\n  \"version\": 1,\n  \"agent\": \"opencode\",\n  \"root_path\": \"/tmp/opencode/skills\",\n  \"skills\": [{\"source_id\": \"source\", \"relative_path\": \"reviewer\"}]\n}\n"), 0o644))

		got, err := repository.LoadAll()

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "opencode", got[0].Adapter())
		assert.Equal(t, "/tmp/opencode/skills", got[0].RootPath())
		require.Len(t, got[0].Identities(), 1)
		assert.Equal(t, "source", got[0].Identities()[0].SourceID())
	})
}
