package catalog_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileRepository(t *testing.T) {
	t.Run("return error when path is empty", func(t *testing.T) {
		_, err := catalog.NewFileRepository("")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "catalog path required")
	})
}

func TestFileRepository(t *testing.T) {
	var (
		newRepository = func(t *testing.T) (*catalog.FileRepository, string) {
			path := filepath.Join(t.TempDir(), "catalog.json")
			repository, err := catalog.NewFileRepository(path)
			require.NoError(t, err)
			return repository, path
		}
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
	)

	t.Run("load zero catalog when file does not exist", func(t *testing.T) {
		repository, _ := newRepository(t)

		got, err := repository.Load()

		require.NoError(t, err)
		assert.True(t, got.IndexedAt().IsZero())
		assert.Nil(t, got.Skills())
	})

	t.Run("save and load catalog snapshot", func(t *testing.T) {
		var (
			repository, _   = newRepository(t)
			indexedAt       = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			reviewerSkill   = newSkill(t, "source-a", "reviewer", "Reviewer")
			programmerSkill = newSkill(t, "source-b", "programmer", "Programmer")
		)

		err := repository.Save(catalog.NewCatalog(indexedAt, reviewerSkill, programmerSkill))

		require.NoError(t, err)

		got, err := repository.Load()
		require.NoError(t, err)
		assert.Equal(t, indexedAt, got.IndexedAt())
		assert.Equal(t, catalog.Skills{reviewerSkill, programmerSkill}, got.Skills())
	})

	t.Run("normalize duplicate entries from file", func(t *testing.T) {
		var (
			repository, path = newRepository(t)
			indexedAt        = time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
			reviewerSkill    = newSkill(t, "source-a", "reviewer", "Reviewer")
		)

		err := os.WriteFile(path, []byte(`{
  "version": 1,
  "generated_at": "2026-03-18T12:00:00Z",
  "skills": [
    {
      "source_id": "source-a",
      "name": "Reviewer",
      "description": "Reviewer description",
      "relative_path": "reviewer",
      "skill_file": "reviewer/SKILL.md"
    },
    {
      "source_id": "source-a",
      "name": "Reviewer",
      "description": "Reviewer description",
      "relative_path": "reviewer",
      "skill_file": "reviewer/SKILL.md"
    }
  ]
}
`), 0o644)
		require.NoError(t, err)

		got, err := repository.Load()
		require.NoError(t, err)
		assert.Equal(t, indexedAt, got.IndexedAt())
		assert.Equal(t, catalog.Skills{reviewerSkill}, got.Skills())
	})
}
