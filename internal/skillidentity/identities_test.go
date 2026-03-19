package skillidentity_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentities(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("sort and deduplicate identities", func(t *testing.T) {
		var (
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			testerIdentity     = newIdentity(t, "source-a", "tester")
		)

		got := skillidentity.NewIdentities(programmerIdentity, reviewerIdentity, testerIdentity, reviewerIdentity)

		assert.Equal(t, skillidentity.Identities{reviewerIdentity, testerIdentity, programmerIdentity}, got)
	})
}

func TestIdentitiesAdd(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("add identity to collection", func(t *testing.T) {
		var (
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			sut                = skillidentity.NewIdentities(reviewerIdentity)
		)

		got, err := sut.Add(programmerIdentity)

		require.NoError(t, err)
		assert.Equal(t, skillidentity.Identities{reviewerIdentity, programmerIdentity}, got)
	})

	t.Run("return error when identity already exists", func(t *testing.T) {
		var (
			reviewerIdentity = newIdentity(t, "source-a", "reviewer")
			sut              = skillidentity.NewIdentities(reviewerIdentity)
		)

		_, err := sut.Add(reviewerIdentity)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestIdentitiesRemove(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("remove identity from collection", func(t *testing.T) {
		var (
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			sut                = skillidentity.NewIdentities(reviewerIdentity, programmerIdentity)
		)

		got, err := sut.Remove(reviewerIdentity)

		require.NoError(t, err)
		assert.Equal(t, skillidentity.Identities{programmerIdentity}, got)
	})

	t.Run("return error when identity is missing", func(t *testing.T) {
		var (
			reviewerIdentity = newIdentity(t, "source-a", "reviewer")
			sut              = skillidentity.NewIdentities()
		)

		_, err := sut.Remove(reviewerIdentity)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
