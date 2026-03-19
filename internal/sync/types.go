package sync

import (
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

// Resolver translates a skill identity into its local source directory.
type Resolver func(skill_identity.Identity) (string, error)

// Result is the full reconciliation report across all sync targets.
type Result struct {
	DesiredCount int
	Targets      []TargetResult
	Manifests    []Manifest
}

// TargetResult is the reconciliation report for one sync destination.
type TargetResult struct {
	Adapter   string
	Adapters  []string
	RootPath  string
	Linked    int
	Removed   int
	Skipped   int
	Unchanged int
	Error     string
}

// Summary returns a short human-readable description of the sync result.
func (r Result) Summary() string {
	var errorCount int
	for _, target := range r.Targets {
		if target.Error != "" {
			errorCount++
		}
	}

	var summary string
	if r.DesiredCount == 0 {
		summary = fmt.Sprintf("Cleared synced skills from %d %s", len(r.Targets), pluralize(len(r.Targets), "location", "locations"))
	} else {
		summary = fmt.Sprintf("Synced %d selected %s to %d %s", r.DesiredCount, pluralize(r.DesiredCount, "skill", "skills"), len(r.Targets), pluralize(len(r.Targets), "location", "locations"))
	}

	if errorCount > 0 {
		summary += fmt.Sprintf(" • %d %s", errorCount, pluralize(errorCount, "error", "errors"))
	}

	return summary
}

// DisplayName returns the most helpful identifier for one target result.
func (r TargetResult) DisplayName() string {
	if r.RootPath == "" {
		return r.Adapter
	}

	if len(r.Adapters) == 0 {
		return r.RootPath
	}

	return fmt.Sprintf("%s (%s)", r.RootPath, strings.Join(r.Adapters, ", "))
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}
