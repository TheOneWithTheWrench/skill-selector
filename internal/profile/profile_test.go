package profile_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/profile"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
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
		assert.Equal(t, skillidentity.NewIdentities(firstIdentity, secondIdentity), got.Selected())
		assert.Equal(t, 2, got.SelectedCount())
	})

	t.Run("return error for empty profile name", func(t *testing.T) {
		_, err := profile.New("   ")

		require.Error(t, err)
		assert.EqualError(t, err, "profile name required")
	})
}

func newIdentity(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
	t.Helper()

	identity, err := skillidentity.New(sourceID, relativePath)
	require.NoError(t, err)

	return identity
}
