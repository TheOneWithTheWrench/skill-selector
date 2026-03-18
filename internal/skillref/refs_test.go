package skillref_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRefs(t *testing.T) {
	var (
		newRef = func(t *testing.T, sourceID string, relativePath string) skillref.Ref {
			ref, err := skillref.New(sourceID, relativePath)
			require.NoError(t, err)
			return ref
		}
	)

	t.Run("sort and deduplicate refs", func(t *testing.T) {
		var (
			reviewerRef   = newRef(t, "source-a", "reviewer")
			programmerRef = newRef(t, "source-b", "programmer")
			testerRef     = newRef(t, "source-a", "tester")
		)

		got := skillref.NewRefs(programmerRef, reviewerRef, testerRef, reviewerRef)

		assert.Equal(t, skillref.Refs{reviewerRef, testerRef, programmerRef}, got)
	})
}

func TestRefsAdd(t *testing.T) {
	var (
		newRef = func(t *testing.T, sourceID string, relativePath string) skillref.Ref {
			ref, err := skillref.New(sourceID, relativePath)
			require.NoError(t, err)
			return ref
		}
	)

	t.Run("add ref to collection", func(t *testing.T) {
		var (
			reviewerRef   = newRef(t, "source-a", "reviewer")
			programmerRef = newRef(t, "source-b", "programmer")
			sut           = skillref.NewRefs(reviewerRef)
		)

		got, err := sut.Add(programmerRef)

		require.NoError(t, err)
		assert.Equal(t, skillref.Refs{reviewerRef, programmerRef}, got)
	})

	t.Run("return error when ref already exists", func(t *testing.T) {
		var (
			reviewerRef = newRef(t, "source-a", "reviewer")
			sut         = skillref.NewRefs(reviewerRef)
		)

		_, err := sut.Add(reviewerRef)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestRefsRemove(t *testing.T) {
	var (
		newRef = func(t *testing.T, sourceID string, relativePath string) skillref.Ref {
			ref, err := skillref.New(sourceID, relativePath)
			require.NoError(t, err)
			return ref
		}
	)

	t.Run("remove ref from collection", func(t *testing.T) {
		var (
			reviewerRef   = newRef(t, "source-a", "reviewer")
			programmerRef = newRef(t, "source-b", "programmer")
			sut           = skillref.NewRefs(reviewerRef, programmerRef)
		)

		got, err := sut.Remove(reviewerRef)

		require.NoError(t, err)
		assert.Equal(t, skillref.Refs{programmerRef}, got)
	})

	t.Run("return error when ref is missing", func(t *testing.T) {
		var (
			reviewerRef = newRef(t, "source-a", "reviewer")
			sut         = skillref.NewRefs()
		)

		_, err := sut.Remove(reviewerRef)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
