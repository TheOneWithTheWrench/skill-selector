package source

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheOneWithTheWrench/skill-selector/internal/file_util"
)

// Refresher updates local mirrors so they match their configured upstream sources.
type Refresher interface {
	Refresh(ctx context.Context, mirror Mirror) (RefreshResult, error)
}

// RefreshResult describes how a mirror was updated.
type RefreshResult struct {
	Mirror Mirror
	Action string
}

// GitRefresher materializes and updates mirrors through git commands.
type GitRefresher struct {
	runner Runner
}

// NewGitRefresher constructs a git-backed refresher for source mirrors.
func NewGitRefresher(runner Runner) (*GitRefresher, error) {
	if runner == nil {
		return nil, fmt.Errorf("source runner required")
	}

	return &GitRefresher{runner: runner}, nil
}

// Refresh clones a missing mirror or fast-forwards an existing one.
func (r GitRefresher) Refresh(ctx context.Context, mirror Mirror) (RefreshResult, error) {
	_, err := os.Stat(mirror.ClonePath)
	if errors.Is(err, os.ErrNotExist) {
		if err := file_util.EnsureDir(filepath.Dir(mirror.ClonePath)); err != nil {
			return RefreshResult{}, err
		}

		if err := r.runner.Run(
			ctx,
			"",
			"git",
			"clone",
			"--branch",
			mirror.Source.Ref(),
			"--single-branch",
			mirror.Source.CloneURL(),
			mirror.ClonePath,
		); err != nil {
			return RefreshResult{}, fmt.Errorf("clone source %q: %w", mirror.Source.Locator(), err)
		}

		return RefreshResult{
			Mirror: mirror,
			Action: "cloned",
		}, nil
	}
	if err != nil {
		return RefreshResult{}, fmt.Errorf("stat source clone %q: %w", mirror.ClonePath, err)
	}

	if err := r.runner.Run(ctx, mirror.ClonePath, "git", "pull", "--ff-only"); err != nil {
		return RefreshResult{}, fmt.Errorf("pull source %q: %w", mirror.Source.Locator(), err)
	}

	return RefreshResult{
		Mirror: mirror,
		Action: "pulled",
	}, nil
}
