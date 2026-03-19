package app_test

import (
	"errors"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	t.Run("list profiles", func(t *testing.T) {
		var (
			profiles = profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer"))
			deps     = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return profiles, nil
		}

		sut := newSut(t, deps)

		got, err := sut.ListProfiles()

		require.NoError(t, err)
		assert.Equal(t, profiles, got)
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		assert.Len(t, deps.ProfileRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("create profile", func(t *testing.T) {
		var (
			storedProfiles = profile.DefaultProfiles()
			deps           = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return storedProfiles, nil
		}
		deps.ProfileRepository.SaveFunc = func(next profile.Profiles) error {
			storedProfiles = next
			return nil
		}

		sut := newSut(t, deps)

		got, err := sut.CreateProfile("reviewer")

		require.NoError(t, err)
		_, ok := got.Find("reviewer")
		assert.True(t, ok)
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("rename profile", func(t *testing.T) {
		var (
			storedProfiles = profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer"))
			deps           = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return storedProfiles, nil
		}
		deps.ProfileRepository.SaveFunc = func(next profile.Profiles) error {
			storedProfiles = next
			return nil
		}

		sut := newSut(t, deps)

		got, err := sut.RenameProfile("reviewer", "editor")

		require.NoError(t, err)
		_, ok := got.Find("editor")
		assert.True(t, ok)
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("switch profile", func(t *testing.T) {
		var (
			storedProfiles = profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer"))
			deps           = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return storedProfiles, nil
		}
		deps.ProfileRepository.SaveFunc = func(next profile.Profiles) error {
			storedProfiles = next
			return nil
		}

		sut := newSut(t, deps)

		got, err := sut.SwitchProfile("reviewer")

		require.NoError(t, err)
		assert.Equal(t, "reviewer", got.ActiveName())
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("remove profile", func(t *testing.T) {
		var (
			storedProfiles = profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer"))
			deps           = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return storedProfiles, nil
		}
		deps.ProfileRepository.SaveFunc = func(next profile.Profiles) error {
			storedProfiles = next
			return nil
		}

		sut := newSut(t, deps)

		got, err := sut.RemoveProfile("reviewer")

		require.NoError(t, err)
		_, ok := got.Find("reviewer")
		assert.False(t, ok)
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("save active profile selection", func(t *testing.T) {
		var (
			identity       = newIdentity(t, "source-a", "reviewer")
			storedProfiles = profile.DefaultProfiles()
			deps           = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return storedProfiles, nil
		}
		deps.ProfileRepository.SaveFunc = func(next profile.Profiles) error {
			storedProfiles = next
			return nil
		}

		sut := newSut(t, deps)

		got, err := sut.SaveActiveProfileSelection(skill_identity.NewIdentities(identity))

		require.NoError(t, err)
		assert.Equal(t, skill_identity.NewIdentities(identity), got.Active().Selected())
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Equal(t, skill_identity.NewIdentities(identity), deps.ProfileRepository.SaveCalls()[0].Profiles.Active().Selected())
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})

	t.Run("return repository save error while switching profile", func(t *testing.T) {
		var (
			expectedErr = errors.New("save failed")
			deps        = newDefaultDependencies()
		)

		deps.ProfileRepository.LoadFunc = func() (profile.Profiles, error) {
			return profile.NewProfiles(profile.DefaultName, profile.Default(), mustProfile(t, "reviewer")), nil
		}
		deps.ProfileRepository.SaveFunc = func(profile.Profiles) error {
			return expectedErr
		}

		sut := newSut(t, deps)

		_, err := sut.SwitchProfile("reviewer")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		require.Len(t, deps.ProfileRepository.LoadCalls(), 1)
		require.Len(t, deps.ProfileRepository.SaveCalls(), 1)
		assert.Len(t, deps.SourceRepository.LoadCalls(), 0)
		assert.Len(t, deps.SourceRepository.SaveCalls(), 0)
		assert.Len(t, deps.SourceRefresher.RefreshCalls(), 0)
		assert.Len(t, deps.CatalogRepository.LoadCalls(), 0)
		assert.Len(t, deps.CatalogRepository.SaveCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.LoadAllCalls(), 0)
		assert.Len(t, deps.SyncManifestRepo.SaveCalls(), 0)
		assert.Len(t, deps.Clock.NowCalls(), 0)
	})
}

func mustProfile(t *testing.T, name string, identities ...skill_identity.Identity) profile.Profile {
	t.Helper()

	item, err := profile.New(name, identities...)
	require.NoError(t, err)

	return item
}
