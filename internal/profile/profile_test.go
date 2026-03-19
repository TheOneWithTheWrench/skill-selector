package profile_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("normalize profile name and selected identities", func(t *testing.T) {
		var (
			firstIdentity  = newIdentity(t, "source-a", "reviewer")
			secondIdentity = newIdentity(t, "source-a", "writer")
		)

		got, err := profile.New(" reviewer ", secondIdentity, firstIdentity, secondIdentity)

		require.NoError(t, err)
		assert.Equal(t, "reviewer", got.Name())
		assert.Equal(t, skill_identity.NewIdentities(firstIdentity, secondIdentity), got.Selected())
		assert.Equal(t, 2, got.SelectedCount())
	})

	t.Run("return error for empty profile name", func(t *testing.T) {
		_, err := profile.New("   ")

		require.Error(t, err)
		assert.EqualError(t, err, "profile name required")
	})

	t.Run("remove identities from one source", func(t *testing.T) {
		var (
			firstIdentity  = newIdentity(t, "source-a", "reviewer")
			secondIdentity = newIdentity(t, "source-b", "writer")
			thirdIdentity  = newIdentity(t, "source-a", "editor")
			item           = mustProfile(t, "reviewer", firstIdentity, secondIdentity, thirdIdentity)
		)

		got := item.WithoutSource("source-a")

		assert.Equal(t, "reviewer", got.Name())
		assert.Equal(t, skill_identity.NewIdentities(secondIdentity), got.Selected())
	})
}

func newIdentity(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
	t.Helper()

	identity, err := skill_identity.New(sourceID, relativePath)
	require.NoError(t, err)

	return identity
}
