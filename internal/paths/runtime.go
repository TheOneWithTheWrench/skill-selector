package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// AppName is the XDG application name used when deriving runtime paths.
	AppName         = "skill-switcher"
	sourcesFileName = "sources.json"
	cacheEnv        = "XDG_CACHE_HOME"
	dataEnv         = "XDG_DATA_HOME"
)

// Runtime groups the XDG-derived paths the application reads and writes during execution.
type Runtime struct {
	CacheDir     string
	DataDir      string
	SourcesFile  string
	SourcesDir   string
	CatalogFile  string
	ProfilesFile string
	SyncStateDir string
	LogsDir      string
}

// Default resolves XDG base paths and derives the runtime file locations from them.
func Default() (Runtime, error) {
	cacheDir, err := resolveDir(cacheEnv, ".cache", AppName)
	if err != nil {
		return Runtime{}, err
	}

	dataDir, err := resolveDir(dataEnv, filepath.Join(".local", "share"), AppName)
	if err != nil {
		return Runtime{}, err
	}

	return Runtime{
		CacheDir:     cacheDir,
		DataDir:      dataDir,
		SourcesFile:  filepath.Join(dataDir, sourcesFileName),
		SourcesDir:   filepath.Join(dataDir, "sources"),
		CatalogFile:  filepath.Join(cacheDir, "catalog.json"),
		ProfilesFile: filepath.Join(dataDir, "profiles.json"),
		SyncStateDir: filepath.Join(dataDir, "activations"),
		LogsDir:      filepath.Join(cacheDir, "logs"),
	}, nil
}

// EnsureRuntimeDirs creates the directories needed before the app reads or writes state.
func (r Runtime) EnsureRuntimeDirs() error {
	var dirs = []string{
		r.CacheDir,
		r.DataDir,
		r.SourcesDir,
		r.SyncStateDir,
		r.LogsDir,
	}

	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			return fmt.Errorf("runtime directory required")
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %q: %w", dir, err)
		}
	}

	return nil
}

func resolveDir(envKey string, fallbackSuffix string, appName string) (string, error) {
	if strings.TrimSpace(appName) == "" {
		return "", fmt.Errorf("app name required")
	}

	if baseDir := strings.TrimSpace(os.Getenv(envKey)); baseDir != "" {
		return filepath.Join(baseDir, appName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, fallbackSuffix, appName), nil
}
