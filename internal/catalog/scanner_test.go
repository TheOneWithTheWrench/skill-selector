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
		newSubtreeMirror = func(t *testing.T, clonePath string, locator string) source.Mirror {
			t.Helper()

			configuredSource, err := source.Parse(locator)
			require.NoError(t, err)

			return source.Mirror{
				Source:    configuredSource,
				ClonePath: clonePath,
			}
		}
		newRootMirror = func(t *testing.T, clonePath string) source.Mirror {
			t.Helper()

			configuredSource, err := source.Parse("https://github.com/ComposioHQ/awesome-claude-skills")
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

	t.Run("parse multiline yaml frontmatter descriptions", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "smart-reporting", "---\nname: report\ndescription: >-\n  Generate test report. Use when user says \"test report\", \"results summary\",\n  \"test status\", \"show results\", or \"how did tests go\".\n---\n\n# Smart Test Reporting\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Smart Test Reporting", skills[0].Name())
		assert.Equal(t, "Generate test report. Use when user says \"test report\", \"results summary\", \"test status\", \"show results\", or \"how did tests go\".", skills[0].Description())
	})

	t.Run("parse yaml frontmatter tags lists", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "smart-reporting", "---\nname: report\ndescription: Generate reports\ntags:\n  - reporting\n  - playwright\n  - Reporting\n---\n\n# Smart Test Reporting\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, []string{"reporting", "playwright"}, skills[0].Tags())
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

	t.Run("discover nested skills and skip the collection root skill", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "engineering-team")
			mirror     = newSubtreeMirror(t, rootPath, "https://github.com/alirezarezvani/claude-skills/tree/main/engineering-team")
		)

		newSkill(t, sourceRoot, "", "# Engineering Team\n\nBundle description.\n")
		newSkill(t, sourceRoot, "a11y-audit", "# A11y Audit\n\nAccessibility checks.\n")
		newSkill(t, sourceRoot, "code-reviewer", "# Code Reviewer\n\nReview code changes.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 2)
		assert.Equal(t, "A11y Audit", skills[0].Name())
		assert.Equal(t, "a11y-audit", skills[0].RelativePath())
		assert.Equal(t, "code-reviewer", skills[1].RelativePath())
	})

	t.Run("discover a root skill from a plain github repo source", func(t *testing.T) {
		var (
			rootPath = t.TempDir()
			mirror   = newRootMirror(t, rootPath)
		)

		newSkill(t, rootPath, "", "# Brand Guidelines\n\nApply brand colors.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Brand Guidelines", skills[0].Name())
		assert.Equal(t, "Apply brand colors.", skills[0].Description())
		assert.Empty(t, skills[0].RelativePath())
	})

	t.Run("skip common heavy directories even if they contain skill files", func(t *testing.T) {
		var (
			rootPath   = t.TempDir()
			sourceRoot = filepath.Join(rootPath, "skills")
			mirror     = newMirror(t, rootPath)
		)

		newSkill(t, sourceRoot, "reviewer", "# Reviewer\n\nReview pull requests carefully.\n")
		newSkill(t, sourceRoot, "node_modules/fake-package", "# Fake Package\n\nShould be ignored.\n")
		newSkill(t, sourceRoot, "vendor/third-party", "# Third Party\n\nShould be ignored.\n")

		skills, err := catalog.Scan(mirror)

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "Reviewer", skills[0].Name())
		assert.Equal(t, "reviewer", skills[0].RelativePath())
	})
}
