package tui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/require"
)

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

func buildSnapshot(runtime paths.Runtime, configuredSources source.Sources, discoveredSkills []catalog.Skill, manifests []skillsync.Manifest) Snapshot {
	return newSnapshot(
		runtime,
		configuredSources,
		catalog.NewCatalog(time.Date(2026, time.March, 19, 12, 0, 0, 0, time.UTC), discoveredSkills...),
		manifests,
	)
}
