package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	t.Run("migrate legacy runtime dirs into skill-selector paths", func(t *testing.T) {
		var (
			rootDir        = t.TempDir()
			cacheRootDir   = filepath.Join(rootDir, "cache")
			dataRootDir    = filepath.Join(rootDir, "data")
			legacyCacheDir = filepath.Join(cacheRootDir, legacyAppName)
			legacyDataDir  = filepath.Join(dataRootDir, legacyAppName)
		)

		t.Setenv(cacheEnv, cacheRootDir)
		t.Setenv(dataEnv, dataRootDir)

		require.NoError(t, os.MkdirAll(filepath.Join(legacyDataDir, "sources"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(legacyCacheDir, "logs"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(legacyDataDir, "profiles.json"), []byte("profiles"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(legacyDataDir, "sources.json"), []byte("sources"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(legacyCacheDir, "catalog.json"), []byte("catalog"), 0o644))

		runtime, err := Default()

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cacheRootDir, AppName), runtime.CacheDir)
		assert.Equal(t, filepath.Join(dataRootDir, AppName), runtime.DataDir)
		assertFileContent(t, filepath.Join(runtime.DataDir, "profiles.json"), "profiles")
		assertFileContent(t, filepath.Join(runtime.DataDir, "sources.json"), "sources")
		assertFileContent(t, filepath.Join(runtime.CacheDir, "catalog.json"), "catalog")
		assert.False(t, pathExists(filepath.Join(dataRootDir, legacyAppName)))
		assert.False(t, pathExists(filepath.Join(cacheRootDir, legacyAppName)))
	})
}

func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
