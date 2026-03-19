package file_util

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates a directory path and any missing parents.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", path, err)
	}

	return nil
}

// EnsureParentDir creates the parent directory for a file path.
func EnsureParentDir(path string) error {
	return EnsureDir(filepath.Dir(path))
}

// WriteFile writes a file atomically by renaming a synced temp file into place.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".skill-selector-*")
	if err != nil {
		return fmt.Errorf("create temp file for %q: %w", path, err)
	}

	tempPath := file.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temp file for %q: %w", path, err)
	}

	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync temp file for %q: %w", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close temp file for %q: %w", path, err)
	}

	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("chmod temp file for %q: %w", path, err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename temp file for %q: %w", path, err)
	}

	return nil
}
