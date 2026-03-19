package profile_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	t.Run("default profiles keeps the default profile active", func(t *testing.T) {
		got := profile.DefaultProfiles()

		assert.Equal(t, profile.DefaultName, got.ActiveName())
		require.Len(t, got.All(), 1)
		assert.Equal(t, profile.DefaultName, got.All()[0].Name())
	})

	t.Run("create keeps default first and sorts other profiles", func(t *testing.T) {
		var (
			profiles = profile.DefaultProfiles()
			err      error
		)

		profiles, err = profiles.Create("zeta")
		require.NoError(t, err)
		profiles, err = profiles.Create("alpha")
		require.NoError(t, err)

		items := profiles.All()
		require.Len(t, items, 3)
		assert.Equal(t, profile.DefaultName, items[0].Name())
		assert.Equal(t, "alpha", items[1].Name())
		assert.Equal(t, "zeta", items[2].Name())
		assert.Equal(t, profile.DefaultName, profiles.ActiveName())
	})

	t.Run("rename updates the active profile name", func(t *testing.T) {
		var (
			profiles = profile.NewProfiles(
				"reviewer",
				profile.Default(),
				mustProfile(t, "reviewer"),
			)
		)

		got, err := profiles.Rename("Reviewer", "editor")

		require.NoError(t, err)
		assert.Equal(t, "editor", got.ActiveName())
		_, ok := got.Find("editor")
		assert.True(t, ok)
	})

	t.Run("remove rejects the active profile", func(t *testing.T) {
		var (
			profiles = profile.NewProfiles(
				"reviewer",
				profile.Default(),
				mustProfile(t, "reviewer"),
			)
		)

		_, err := profiles.Remove("reviewer")

		require.Error(t, err)
		assert.EqualError(t, err, "cannot remove active profile: reviewer")
	})

	t.Run("switch changes the active profile without changing selections", func(t *testing.T) {
		var (
			identity = newIdentity(t, "source-a", "reviewer")
			profiles = profile.NewProfiles(
				profile.DefaultName,
				mustProfile(t, profile.DefaultName),
				mustProfile(t, "reviewer", identity),
			)
		)

		got, err := profiles.Switch("reviewer")

		require.NoError(t, err)
		assert.Equal(t, "reviewer", got.ActiveName())
		assert.Equal(t, skill_identity.NewIdentities(identity), got.Active().Selected())
	})

	t.Run("set active selection replaces only the active profile selection", func(t *testing.T) {
		var (
			identity = newIdentity(t, "source-a", "reviewer")
			profiles = profile.NewProfiles(
				"writer",
				profile.Default(),
				mustProfile(t, "writer"),
			)
		)

		got, err := profiles.SetActiveSelection(skill_identity.NewIdentities(identity))

		require.NoError(t, err)
		assert.Equal(t, skill_identity.NewIdentities(identity), got.Active().Selected())
		assert.Empty(t, got.All()[0].Selected())
	})

	t.Run("return error when creating duplicate profile", func(t *testing.T) {
		var (
			profiles = profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer"))
		)

		_, err := profiles.Create("Reviewer")

		require.Error(t, err)
		assert.EqualError(t, err, "profile already exists: Reviewer")
	})

	t.Run("remove one source from every profile selection", func(t *testing.T) {
		var (
			removedIdentity = newIdentity(t, "source-a", "reviewer")
			keptIdentity    = newIdentity(t, "source-b", "writer")
			profiles        = profile.NewProfiles(
				"reviewer",
				mustProfile(t, profile.DefaultName, removedIdentity, keptIdentity),
				mustProfile(t, "reviewer", removedIdentity),
			)
		)

		got := profiles.WithoutSource("source-a")

		assert.Equal(t, "reviewer", got.ActiveName())
		assert.Equal(t, skill_identity.NewIdentities(keptIdentity), got.All()[0].Selected())
		assert.Empty(t, got.Active().Selected())
	})
}

func mustProfile(t *testing.T, name string, identities ...skill_identity.Identity) profile.Profile {
	t.Helper()

	item, err := profile.New(name, identities...)
	require.NoError(t, err)

	return item
}
