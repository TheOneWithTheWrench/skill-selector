package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/agent"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDefinitions(t *testing.T) {
	var (
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			t.Helper()

			identity, err := skillidentity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
	)

	t.Run("resolve shared agents skills defaults for ampcode and codex", func(t *testing.T) {
		definitions := agent.DefaultDefinitions()

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expectedRoot := filepath.Join(homeDir, ".agents", "skills")

		for _, name := range []string{"ampcode", "codex"} {
			definition := lookupDefinition(t, definitions, name)

			rootPath, err := definition.ResolveRoot("")
			require.NoError(t, err)
			assert.Equal(t, expectedRoot, rootPath)
		}
	})

	t.Run("resolve claude default root", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "claude")

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		rootPath, err := definition.ResolveRoot("")

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, ".claude", "skills"), rootPath)
	})

	t.Run("resolve opencode default root", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "opencode")

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		rootPath, err := definition.ResolveRoot("")

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, ".config", "opencode", "skills"), rootPath)
	})

	t.Run("use explicit root override when provided", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "codex")
		rootOverride := filepath.Join(t.TempDir(), "custom-codex-skills")

		rootPath, err := definition.ResolveRoot(rootOverride)

		require.NoError(t, err)
		assert.Equal(t, rootOverride, rootPath)
	})

	t.Run("link path stays within root for normal relative path", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "codex")
		rootPath := "/tmp/agents/skills"
		identity := newIdentity(t, "source", "reviewer")

		target, err := definition.Target(rootPath)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(rootPath, "reviewer"), target.LinkPath(identity))
	})

	t.Run("link path returns root for empty relative path", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "codex")
		rootPath := "/tmp/agents/skills"
		identity := newIdentity(t, "source", "")

		target, err := definition.Target(rootPath)

		require.NoError(t, err)
		assert.Equal(t, rootPath, target.LinkPath(identity))
	})

	t.Run("reject non absolute root overrides", func(t *testing.T) {
		definition := lookupDefinition(t, agent.DefaultDefinitions(), "codex")

		_, err := definition.ResolveRoot("relative/path")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be absolute or start with ~")
	})

	t.Run("build default targets for supported agents", func(t *testing.T) {
		targets, err := agent.DefaultTargets()

		require.NoError(t, err)
		require.Len(t, targets, 4)
		assert.Equal(t, []string{"ampcode", "claude", "codex", "opencode"}, []string{
			targets[0].Adapter(),
			targets[1].Adapter(),
			targets[2].Adapter(),
			targets[3].Adapter(),
		})
	})
}

func lookupDefinition(t *testing.T, definitions []agent.Definition, name string) agent.Definition {
	t.Helper()

	for _, definition := range definitions {
		if definition.Name() == name {
			return definition
		}
	}

	t.Fatalf("definition not found: %s", name)
	return agent.Definition{}
}
