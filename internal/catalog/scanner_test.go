package catalog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	var (
		newSkill = func(t *testing.T, rootPath string, relativePath string, content string) {
			t.Helper()

			skillDir := filepath.Join(rootPath, filepath.FromSlash(relativePath))
			require.NoError(t, os.MkdirAll(skillDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644))
		}
		newMirror = func(t *testing.T, clonePath string) source.Mirror {
			t.Helper()

			configuredSource, err := source.Parse("https://github.com/anthropics/skills/tree/main/skills")
			require.NoError(t, err)

			return source.Mirror{
				Source:    configuredSource,
				ClonePath: clonePath,
			}
		}
	)

	t.Run("discover skills by skill dot md directories", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "reviewer", "# Reviewer\n\nReview pull requests carefully.\n")
		newSkill(t, sourceRoot, "programmer", "# Programmer\n\nImplement changes safely.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 2)
		assert.Equal(t, "Programmer", skills[0].Name())
		assert.Equal(t, "Implement changes safely.", skills[0].Description())
		assert.Equal(t, "reviewer", skills[1].RelativePath())
	})

	t.Run("prefer frontmatter description for skill preview metadata", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "acceptance-testing", "---\nname: writing-acceptance-tests\ndescription: \"End-to-end tests with Given/When/Then pattern.\"\n---\n\n# Acceptance Testing\n\nEnd-to-end tests with real infrastructure.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Acceptance Testing", skills[0].Name())
		assert.Equal(t, "End-to-end tests with Given/When/Then pattern.", skills[0].Description())
	})

	t.Run("prefer first markdown heading when frontmatter already defines description", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "acceptance-testing", "---\ndescription: Frontmatter description\n---\n\n# Acceptance Testing\n\n# Secondary Heading\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Acceptance Testing", skills[0].Name())
		assert.Equal(t, "Frontmatter description", skills[0].Description())
	})

	t.Run("ignore dangling frontmatter delimiter in description fallback", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "acceptance-testing", "---\nname: writing-acceptance-tests\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "writing-acceptance-tests", skills[0].Name())
		assert.Empty(t, skills[0].Description())
	})

	t.Run("ignore empty frontmatter keys", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "acceptance-testing", "---\nname: Acceptance Testing\n: orphan value\n---\n\nDescription here.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Acceptance Testing", skills[0].Name())
		assert.Equal(t, "Description here.", skills[0].Description())
	})
}
