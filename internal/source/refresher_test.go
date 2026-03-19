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
			runner = &RunnerMock{}
		)

		runner.RunFunc = func(ctx context.Context, workdir string, name string, args ...string) error {
			return os.MkdirAll(mirror.ClonePath, 0o755)
		}

		sut, err := source.NewGitRefresher(runner)
		require.NoError(t, err)

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "cloned", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, runner.RunCalls(), 1)
		assert.Equal(t, "", runner.RunCalls()[0].Workdir)
		assert.Equal(t, "git", runner.RunCalls()[0].Name)
		assert.Equal(t, []string{"clone", "--branch", "main", "--single-branch", "https://github.com/anthropics/skills.git", mirror.ClonePath}, runner.RunCalls()[0].Args)
		assert.DirExists(t, filepath.Dir(mirror.ClonePath))
	})

	t.Run("pull existing mirror", func(t *testing.T) {
		var (
			ctx    = context.Background()
			mirror = newMirror(t)
			runner = &RunnerMock{RunFunc: func(context.Context, string, string, ...string) error { return nil }}
		)

		require.NoError(t, os.MkdirAll(mirror.ClonePath, 0o755))

		sut, err := source.NewGitRefresher(runner)
		require.NoError(t, err)

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "pulled", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, runner.RunCalls(), 1)
		assert.Equal(t, mirror.ClonePath, runner.RunCalls()[0].Workdir)
		assert.Equal(t, "git", runner.RunCalls()[0].Name)
		assert.Equal(t, []string{"pull", "--ff-only"}, runner.RunCalls()[0].Args)
	})

	t.Run("return wrapped clone error", func(t *testing.T) {
		var (
			ctx         = context.Background()
			mirror      = newMirror(t)
			expectedErr = errors.New("clone failed")
			runner      = &RunnerMock{RunFunc: func(ctx context.Context, workdir string, name string, args ...string) error {
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
