package sync

import (
	"errors"
	"fmt"
	"os"

	"github.com/TheOneWithTheWrench/skill-selector/internal/file_util"
)

func ensureSymlink(targetPath string, linkPath string) (bool, error) {
	if err := file_util.EnsureParentDir(linkPath); err != nil {
		return false, err
	}

	linkInfo, err := os.Lstat(linkPath)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Symlink(targetPath, linkPath); err != nil {
			return false, fmt.Errorf("create symlink %q -> %q: %w", linkPath, targetPath, err)
		}

		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("lstat link path %q: %w", linkPath, err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Errorf("refusing to overwrite non-symlink path %q", linkPath)
	}

	existingTarget, err := os.Readlink(linkPath)
	if err != nil {
		return false, fmt.Errorf("read existing symlink %q: %w", linkPath, err)
	}

	if existingTarget == targetPath {
		return false, nil
	}

	if err := os.Remove(linkPath); err != nil {
		return false, fmt.Errorf("remove stale symlink %q: %w", linkPath, err)
	}

	if err := os.Symlink(targetPath, linkPath); err != nil {
		return false, fmt.Errorf("create symlink %q -> %q: %w", linkPath, targetPath, err)
	}

	return true, nil
}

func removeOwnedLink(linkPath string) (bool, error) {
	info, err := os.Lstat(linkPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lstat link path %q: %w", linkPath, err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Errorf("refusing to remove non-symlink path %q", linkPath)
	}

	if err := os.Remove(linkPath); err != nil {
		return false, fmt.Errorf("remove stale symlink %q: %w", linkPath, err)
	}

	return true, nil
}
