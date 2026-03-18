package source_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runnerCall struct {
	workdir string
	name    string
	args    []string
}

type runner struct {
	runFunc func(ctx context.Context, workdir string, name string, args ...string) error
	calls   []runnerCall
}

func (r *runner) Run(ctx context.Context, workdir string, name string, args ...string) error {
	r.calls = append(r.calls, runnerCall{
		workdir: workdir,
		name:    name,
		args:    append([]string(nil), args...),
	})

	if r.runFunc != nil {
		return r.runFunc(ctx, workdir, name, args...)
	}

	return nil
}

func TestNewGitRefresher(t *testing.T) {
	t.Run("return error when runner is missing", func(t *testing.T) {
		_, err := source.NewGitRefresher(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "source runner required")
	})
}

func TestGitRefresherRefresh(t *testing.T) {
	var (
		newMirror = func(t *testing.T) source.Mirror {
			configuredSource, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")
			require.NoError(t, err)

			mirror, err := source.NewMirror(configuredSource, filepath.Join(t.TempDir(), "sources"))
			require.NoError(t, err)
			return mirror
		}
	)

	t.Run("clone missing mirror", func(t *testing.T) {
		var (
			ctx    = context.Background()
			mirror = newMirror(t)
			runner = &runner{}
		)

		runner.runFunc = func(ctx context.Context, workdir string, name string, args ...string) error {
			return os.MkdirAll(mirror.ClonePath, 0o755)
		}

		sut, err := source.NewGitRefresher(runner)
		require.NoError(t, err)

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "cloned", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, runner.calls, 1)
		assert.Equal(t, "", runner.calls[0].workdir)
		assert.Equal(t, "git", runner.calls[0].name)
		assert.Equal(t, []string{"clone", "--branch", "main", "--single-branch", "https://github.com/anthropics/skills.git", mirror.ClonePath}, runner.calls[0].args)
		assert.DirExists(t, filepath.Dir(mirror.ClonePath))
	})

	t.Run("pull existing mirror", func(t *testing.T) {
		var (
			ctx    = context.Background()
			mirror = newMirror(t)
			runner = &runner{}
		)

		require.NoError(t, os.MkdirAll(mirror.ClonePath, 0o755))

		sut, err := source.NewGitRefresher(runner)
		require.NoError(t, err)

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "pulled", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, runner.calls, 1)
		assert.Equal(t, mirror.ClonePath, runner.calls[0].workdir)
		assert.Equal(t, "git", runner.calls[0].name)
		assert.Equal(t, []string{"pull", "--ff-only"}, runner.calls[0].args)
	})

	t.Run("return wrapped clone error", func(t *testing.T) {
		var (
			ctx         = context.Background()
			mirror      = newMirror(t)
			expectedErr = errors.New("clone failed")
			runner      = &runner{runFunc: func(ctx context.Context, workdir string, name string, args ...string) error {
				return expectedErr
			}}
		)

		sut, err := source.NewGitRefresher(runner)
		require.NoError(t, err)

		_, err = sut.Refresh(ctx, mirror)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), mirror.Source.Locator())
	})
}
