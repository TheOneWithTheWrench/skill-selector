package source_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
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
	type dependencies struct {
		Runner *RunnerMock
	}

	var (
		newDefaultDependencies = func() *dependencies {
			return &dependencies{
				Runner: &RunnerMock{
					RunFunc: func(context.Context, string, string, ...string) error { return nil },
				},
			}
		}
		newSut = func(t *testing.T, deps *dependencies) *source.GitRefresher {
			t.Helper()

			sut, err := source.NewGitRefresher(deps.Runner)
			require.NoError(t, err)

			return sut
		}
		newCtx = func() context.Context {
			return context.Background()
		}
		newMirror = func(t *testing.T) source.Mirror {
			t.Helper()

			configuredSource, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")
			require.NoError(t, err)

			mirror, err := source.NewMirror(configuredSource, filepath.Join(t.TempDir(), "sources"))
			require.NoError(t, err)

			return mirror
		}
		newRootMirror = func(t *testing.T) source.Mirror {
			t.Helper()

			configuredSource, err := source.Parse("https://github.com/ComposioHQ/awesome-claude-skills")
			require.NoError(t, err)

			mirror, err := source.NewMirror(configuredSource, filepath.Join(t.TempDir(), "sources"))
			require.NoError(t, err)

			return mirror
		}
	)

	t.Run("clone missing mirror", func(t *testing.T) {
		var (
			ctx    = newCtx()
			deps   = newDefaultDependencies()
			mirror = newMirror(t)
			sut    = newSut(t, deps)
		)

		deps.Runner.RunFunc = func(ctx context.Context, workdir string, name string, args ...string) error {
			return os.MkdirAll(mirror.ClonePath, 0o755)
		}

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "cloned", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, deps.Runner.RunCalls(), 1)
		assert.Equal(t, "", deps.Runner.RunCalls()[0].Workdir)
		assert.Equal(t, "git", deps.Runner.RunCalls()[0].Name)
		assert.Equal(t, []string{"clone", "--depth", "1", "--branch", "main", "--single-branch", "https://github.com/anthropics/skills.git", mirror.ClonePath}, deps.Runner.RunCalls()[0].Args)
		assert.DirExists(t, filepath.Dir(mirror.ClonePath))
	})

	t.Run("pull existing mirror", func(t *testing.T) {
		var (
			ctx    = newCtx()
			deps   = newDefaultDependencies()
			mirror = newMirror(t)
			sut    = newSut(t, deps)
		)

		require.NoError(t, os.MkdirAll(mirror.ClonePath, 0o755))

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "pulled", got.Action)
		assert.Equal(t, mirror, got.Mirror)
		require.Len(t, deps.Runner.RunCalls(), 1)
		assert.Equal(t, mirror.ClonePath, deps.Runner.RunCalls()[0].Workdir)
		assert.Equal(t, "git", deps.Runner.RunCalls()[0].Name)
		assert.Equal(t, []string{"pull", "--ff-only"}, deps.Runner.RunCalls()[0].Args)
	})

	t.Run("clone repo root source without explicit branch", func(t *testing.T) {
		var (
			ctx    = newCtx()
			deps   = newDefaultDependencies()
			mirror = newRootMirror(t)
			sut    = newSut(t, deps)
		)

		deps.Runner.RunFunc = func(ctx context.Context, workdir string, name string, args ...string) error {
			return os.MkdirAll(mirror.ClonePath, 0o755)
		}

		got, err := sut.Refresh(ctx, mirror)

		require.NoError(t, err)
		assert.Equal(t, "cloned", got.Action)
		require.Len(t, deps.Runner.RunCalls(), 1)
		assert.Equal(t, []string{"clone", "--depth", "1", "https://github.com/ComposioHQ/awesome-claude-skills.git", mirror.ClonePath}, deps.Runner.RunCalls()[0].Args)
	})

	t.Run("return wrapped clone error", func(t *testing.T) {
		var (
			ctx         = newCtx()
			deps        = newDefaultDependencies()
			mirror      = newMirror(t)
			expectedErr = errors.New("clone failed")
			sut         = newSut(t, deps)
		)

		deps.Runner.RunFunc = func(context.Context, string, string, ...string) error {
			return expectedErr
		}

		_, err := sut.Refresh(ctx, mirror)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), mirror.Source.Locator())
		require.Len(t, deps.Runner.RunCalls(), 1)
	})
}
