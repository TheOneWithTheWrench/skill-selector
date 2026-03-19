package skill_identity_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentities(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			identity, err := skill_identity.New(sourceID, relativePath)
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

		got := skill_identity.NewIdentities(programmerIdentity, reviewerIdentity, testerIdentity, reviewerIdentity)

		assert.Equal(t, skill_identity.Identities{reviewerIdentity, testerIdentity, programmerIdentity}, got)
	})
}

func TestIdentitiesAdd(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("add identity to collection", func(t *testing.T) {
		var (
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			sut                = skill_identity.NewIdentities(reviewerIdentity)
		)

		got, err := sut.Add(programmerIdentity)

		require.NoError(t, err)
		assert.Equal(t, skill_identity.Identities{reviewerIdentity, programmerIdentity}, got)
	})

	t.Run("return error when identity already exists", func(t *testing.T) {
		var (
			reviewerIdentity = newIdentity(t, "source-a", "reviewer")
			sut              = skill_identity.NewIdentities(reviewerIdentity)
		)

		_, err := sut.Add(reviewerIdentity)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestIdentitiesRemove(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("remove identity from collection", func(t *testing.T) {
		var (
			reviewerIdentity   = newIdentity(t, "source-a", "reviewer")
			programmerIdentity = newIdentity(t, "source-b", "programmer")
			sut                = skill_identity.NewIdentities(reviewerIdentity, programmerIdentity)
		)

		got, err := sut.Remove(reviewerIdentity)

		require.NoError(t, err)
		assert.Equal(t, skill_identity.Identities{programmerIdentity}, got)
	})

	t.Run("return error when identity is missing", func(t *testing.T) {
		var (
			reviewerIdentity = newIdentity(t, "source-a", "reviewer")
			sut              = skill_identity.NewIdentities()
		)

		_, err := sut.Remove(reviewerIdentity)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
