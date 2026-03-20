package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// Definition describes one built-in sync target for a supported agent.
type Definition struct {
	name        string
	defaultRoot string
}

// DefaultDefinitions returns the built-in sync target definitions we support today.
func DefaultDefinitions() []Definition {
	definitions := []Definition{
		{name: "ampcode", defaultRoot: "~/.agents/skills"},
		{name: "claude", defaultRoot: "~/.claude/skills"},
		{name: "codex", defaultRoot: "~/.agents/skills"},
		{name: "cursor", defaultRoot: "~/.agents/skills"},
		{name: "opencode", defaultRoot: "~/.config/opencode/skills"},
	}

	sort.Slice(definitions, func(left int, right int) bool {
		return definitions[left].name < definitions[right].name
	})

	return definitions
}

// DefaultTargets builds sync targets for the built-in agent definitions.
func DefaultTargets() ([]skillsync.Target, error) {
	definitions := DefaultDefinitions()
	targets := make([]skillsync.Target, 0, len(definitions))

	for _, definition := range definitions {
		target, err := definition.Target("")
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}

// Name returns the stable name used for one built-in agent target.
func (d Definition) Name() string {
	return d.name
}

// ResolveRoot expands the configured root path for the target.
func (d Definition) ResolveRoot(rootOverride string) (string, error) {
	rootPath := strings.TrimSpace(rootOverride)
	if rootPath == "" {
		rootPath = d.defaultRoot
	}

	if rootPath == "~" || strings.HasPrefix(rootPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory for %s: %w", d.name, err)
		}

		if rootPath == "~" {
			rootPath = homeDir
		} else {
			rootPath = filepath.Join(homeDir, strings.TrimPrefix(rootPath, "~/"))
		}
	}

	if !filepath.IsAbs(rootPath) {
		return "", fmt.Errorf("agent root path for %s must be absolute or start with ~: %q", d.name, rootPath)
	}

	return filepath.Clean(rootPath), nil
}

// Target materializes one sync target for the definition.
func (d Definition) Target(rootOverride string) (skillsync.Target, error) {
	rootPath, err := d.ResolveRoot(rootOverride)
	if err != nil {
		return skillsync.Target{}, err
	}

	return skillsync.NewTarget(d.name, rootPath, func(identity skill_identity.Identity) string {
		relativePath := identity.RelativePath()
		if relativePath == "" {
			relativePath = identity.SourceID()
		}

		return safeJoin(rootPath, relativePath)
	})
}

func safeJoin(root string, relativePath string) string {
	cleaned := filepath.Clean(filepath.FromSlash(relativePath))
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return root
	}

	joined := filepath.Join(root, cleaned)
	cleanRoot := filepath.Clean(root)
	rootPrefix := cleanRoot + string(filepath.Separator)
	if joined != cleanRoot && !strings.HasPrefix(joined, rootPrefix) {
		return cleanRoot
	}

	return joined
}
