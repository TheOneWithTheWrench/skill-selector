package catalog_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkills(t *testing.T) {
	var (
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)

			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
	)

	t.Run("sort by source then name and remove duplicate ids", func(t *testing.T) {
		var (
			firstSkill  = newSkill(t, "source-a", "reviewer", "Reviewer")
			secondSkill = newSkill(t, "source-b", "programmer", "Programmer")
			thirdSkill  = newSkill(t, "source-a", "tester", "Tester")
		)

		got := catalog.NewSkills(secondSkill, firstSkill, thirdSkill, firstSkill)

		assert.Equal(t, catalog.Skills{firstSkill, thirdSkill, secondSkill}, got)
	})
}
