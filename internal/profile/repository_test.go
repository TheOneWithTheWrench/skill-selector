package profile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository(t *testing.T) {
	t.Run("load missing file as default profiles", func(t *testing.T) {
		var (
			repository, _ = newRepository(t)
		)

		got, err := repository.Load()

		require.NoError(t, err)
		assert.Equal(t, profile.DefaultProfiles(), got)
	})

	t.Run("save and load current file format", func(t *testing.T) {
		var (
			identity      = newIdentity(t, "source-a", "reviewer")
			repository, _ = newRepository(t)
			profiles      = profile.NewProfiles(
				"reviewer",
				profile.Default(),
				mustProfile(t, "reviewer", identity),
			)
		)

		err := repository.Save(profiles)
		require.NoError(t, err)

		got, err := repository.Load()

		require.NoError(t, err)
		assert.Equal(t, profiles, got)
	})

	t.Run("load legacy default-only file", func(t *testing.T) {
		var (
			identity         = newIdentity(t, "source-a", "reviewer")
			repository, path = newRepository(t)
		)

		legacy := []byte("{\n  \"version\": 1,\n  \"default\": {\n    \"selected_skills\": [\n      {\n        \"source_id\": \"source-a\",\n        \"relative_path\": \"reviewer\"\n      }\n    ]\n  }\n}\n")
		err := os.WriteFile(path, legacy, 0o644)
		require.NoError(t, err)

		got, err := repository.Load()

		require.NoError(t, err)
		assert.Equal(t, profile.DefaultName, got.ActiveName())
		assert.Equal(t, skill_identity.NewIdentities(identity), got.Active().Selected())
	})

	t.Run("return decode error for invalid identity", func(t *testing.T) {
		var (
			repository, path = newRepository(t)
		)

		invalid := []byte("{\n  \"version\": 2,\n  \"active_profile\": \"Default\",\n  \"profiles\": [\n    {\n      \"name\": \"Default\",\n      \"selected_skills\": [\n        {\n          \"source_id\": \"\",\n          \"relative_path\": \"reviewer\"\n        }\n      ]\n    }\n  ]\n}\n")
		err := os.WriteFile(path, invalid, 0o644)
		require.NoError(t, err)

		_, err = repository.Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode profiles file")
	})
}

func newRepository(t *testing.T) (*profile.FileRepository, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "profiles.json")

	repository, err := profile.NewFileRepository(path)
	require.NoError(t, err)

	return repository, path
}
