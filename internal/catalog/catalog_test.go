package catalog_test

import (
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCatalog(t *testing.T) {
	var (
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
	)

	t.Run("normalize indexed at time and discovered skills", func(t *testing.T) {
		var (
			indexedAt       = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
			reviewerSkill   = newSkill(t, "source-a", "reviewer", "Reviewer")
			programmerSkill = newSkill(t, "source-a", "programmer", "Programmer")
		)

		got := catalog.NewCatalog(indexedAt, reviewerSkill, programmerSkill, reviewerSkill)

		assert.Equal(t, indexedAt.UTC(), got.IndexedAt())
		assert.Equal(t, catalog.Skills{programmerSkill, reviewerSkill}, got.Skills())
	})
}

func TestCatalogReplaceSource(t *testing.T) {
	var (
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
	)

	t.Run("replace one source while keeping others", func(t *testing.T) {
		var (
			indexedAt         = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			firstSourceSkill  = newSkill(t, "source-a", "reviewer", "Reviewer")
			secondSourceSkill = newSkill(t, "source-b", "programmer", "Programmer")
			replacementSkill  = newSkill(t, "source-a", "tester", "Tester")
			sut               = catalog.NewCatalog(indexedAt, firstSourceSkill, secondSourceSkill)
		)

		got, err := sut.ReplaceSource("source-a", catalog.Skills{replacementSkill})

		require.NoError(t, err)
		assert.Equal(t, indexedAt, got.IndexedAt())
		assert.Equal(t, catalog.Skills{replacementSkill, secondSourceSkill}, got.Skills())
		assert.Equal(t, catalog.Skills{firstSourceSkill, secondSourceSkill}, sut.Skills())
	})

	t.Run("remove source skills", func(t *testing.T) {
		var (
			firstSourceSkill  = newSkill(t, "source-a", "reviewer", "Reviewer")
			secondSourceSkill = newSkill(t, "source-b", "programmer", "Programmer")
			sut               = catalog.NewCatalog(time.Time{}, firstSourceSkill, secondSourceSkill)
		)

		got, err := sut.RemoveSource("source-a")

		require.NoError(t, err)
		assert.Equal(t, catalog.Skills{secondSourceSkill}, got.Skills())
	})

	t.Run("return error when source id is missing", func(t *testing.T) {
		sut := catalog.NewCatalog(time.Time{})

		_, err := sut.ReplaceSource("", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog source id required")
	})

	t.Run("return error when replacement skill belongs to another source", func(t *testing.T) {
		var (
			foreignSkill = newSkill(t, "source-b", "programmer", "Programmer")
			sut          = catalog.NewCatalog(time.Time{})
		)

		_, err := sut.ReplaceSource("source-a", catalog.Skills{foreignSkill})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog replacement skill source mismatch")
	})
}
