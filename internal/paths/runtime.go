package paths

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// AppName is the XDG application name used when deriving runtime paths.
	AppName         = "skill-selector"
	legacyAppName   = "skill-switcher"
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
	runtime, err := defaultRuntime(AppName)
	if err != nil {
		return Runtime{}, err
	}

	legacyRuntime, err := defaultRuntime(legacyAppName)
	if err != nil {
		return Runtime{}, err
	}

	if err := migrateLegacyRuntime(legacyRuntime, runtime); err != nil {
		return Runtime{}, err
	}

	return runtime, nil
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

func defaultRuntime(appName string) (Runtime, error) {
	cacheDir, err := resolveDir(cacheEnv, ".cache", appName)
	if err != nil {
		return Runtime{}, err
	}

	dataDir, err := resolveDir(dataEnv, filepath.Join(".local", "share"), appName)
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

func migrateLegacyRuntime(legacy Runtime, current Runtime) error {
	if err := migrateDir(legacy.DataDir, current.DataDir); err != nil {
		return fmt.Errorf("migrate data dir: %w", err)
	}
	if err := migrateDir(legacy.CacheDir, current.CacheDir); err != nil {
		return fmt.Errorf("migrate cache dir: %w", err)
	}

	return nil
}

func migrateDir(oldDir string, newDir string) error {
	if strings.TrimSpace(oldDir) == "" || strings.TrimSpace(newDir) == "" || oldDir == newDir {
		return nil
	}

	if err := ensureTargetMissingOrUsable(newDir); err != nil {
		return err
	}
	if exists, err := dirExists(newDir); err != nil {
		return fmt.Errorf("stat current dir %q: %w", newDir, err)
	} else if exists {
		return nil
	}

	if exists, err := dirExists(oldDir); err != nil {
		return fmt.Errorf("stat legacy dir %q: %w", oldDir, err)
	} else if !exists {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(newDir), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %q: %w", newDir, err)
	}

	if err := os.Rename(oldDir, newDir); err == nil {
		return nil
	}

	if err := copyDir(oldDir, newDir); err != nil {
		return fmt.Errorf("copy legacy dir %q to %q: %w", oldDir, newDir, err)
	}
	if err := os.RemoveAll(oldDir); err != nil {
		return fmt.Errorf("remove legacy dir %q: %w", oldDir, err)
	}

	return nil
}

func ensureTargetMissingOrUsable(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("target path %q is not a directory", dir)
	}

	return nil
}

func dirExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func copyDir(sourceDir string, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := targetDir
		if relativePath != "." {
			targetPath = filepath.Join(targetDir, relativePath)
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		return copyFile(path, targetPath, info.Mode().Perm())
	})
}

func copyFile(sourcePath string, targetPath string, perm fs.FileMode) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	return nil
}
