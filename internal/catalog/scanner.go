package catalog

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	"gopkg.in/yaml.v3"
)

type scanCandidate struct {
	relativePath string
	skill        Skill
}

type frontmatterMetadata struct {
	Name        string
	Description string
	Tags        []string
}

var skippedDirectoryNames = map[string]struct{}{
	".git":         {},
	"__pycache__":  {},
	"build":        {},
	"coverage":     {},
	"dist":         {},
	"node_modules": {},
	"out":          {},
	"target":       {},
	"tmp":          {},
	"vendor":       {},
	"venv":         {},
}

// Scan walks a source mirror and discovers skill directories backed by `SKILL.md` files.
// When one directory contains both a top-level `SKILL.md` and nested skill directories,
// the nested skills win and the parent directory is treated as a collection instead.
func Scan(mirror source.Mirror) (Skills, error) {
	rootPath := mirror.SubtreePath()

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("stat source subtree %q: %w", rootPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source subtree is not a directory: %q", rootPath)
	}

	var candidates []scanCandidate
	err = filepath.WalkDir(rootPath, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !entry.IsDir() {
			return nil
		}

		if shouldSkipDirectory(entry.Name(), currentPath, rootPath) {
			return filepath.SkipDir
		}

		skillFilePath := filepath.Join(currentPath, "SKILL.md")
		if _, err := os.Stat(skillFilePath); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(rootPath, currentPath)
		if err != nil {
			return fmt.Errorf("compute relative path for %q: %w", currentPath, err)
		}

		relativePath = filepath.ToSlash(relativePath)
		if relativePath == "." {
			relativePath = ""
		}

		name, description, tags, err := readSkillMetadata(skillFilePath, filepath.Base(currentPath))
		if err != nil {
			return err
		}

		identity, err := skill_identity.New(mirror.ID(), relativePath)
		if err != nil {
			return err
		}

		discoveredSkill, err := NewSkill(identity, name, description, tags...)
		if err != nil {
			return err
		}

		candidates = append(candidates, scanCandidate{
			relativePath: relativePath,
			skill:        discoveredSkill,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan source %q: %w", mirror.ID(), err)
	}

	return NewSkills(filterScanCandidates(candidates)...), nil
}

func filterScanCandidates(candidates []scanCandidate) []Skill {
	filtered := make([]Skill, 0, len(candidates))

	for _, candidate := range candidates {
		if hasNestedCandidate(candidate.relativePath, candidates) {
			continue
		}

		filtered = append(filtered, candidate.skill)
	}

	return filtered
}

func hasNestedCandidate(relativePath string, candidates []scanCandidate) bool {
	for _, candidate := range candidates {
		if candidate.relativePath == relativePath {
			continue
		}

		if relativePath == "" {
			return true
		}

		if strings.HasPrefix(candidate.relativePath, relativePath+"/") {
			return true
		}
	}

	return false
}

func shouldSkipDirectory(name string, currentPath string, rootPath string) bool {
	if currentPath == rootPath {
		return false
	}

	if _, ok := skippedDirectoryNames[name]; ok {
		return true
	}

	return strings.HasPrefix(name, ".")
}

func readSkillMetadata(path string, fallbackName string) (string, string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", nil, fmt.Errorf("open skill file %q: %w", path, err)
	}
	defer file.Close()

	var (
		lines       []string
		name        = fallbackName
		description string
		scanner     = bufio.NewScanner(file)
	)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", "", nil, fmt.Errorf("scan skill file %q: %w", path, err)
	}

	startIndex := 0
	nameFromHeading := false
	var tags []string
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		frontmatter, nextIndex := readFrontmatter(lines)
		startIndex = nextIndex

		if frontmatter.Name != "" {
			name = frontmatter.Name
		}
		if frontmatter.Description != "" {
			description = frontmatter.Description
		}
		tags = append(tags, frontmatter.Tags...)
	}

	for _, rawLine := range lines[startIndex:] {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if after, ok := strings.CutPrefix(line, "# "); ok {
			if !nameFromHeading {
				name = strings.TrimSpace(after)
				nameFromHeading = true
			}
			continue
		}

		if description == "" && !strings.HasPrefix(line, "#") {
			description = line
		}
	}

	return name, description, normalizeTags(tags...), nil
}

func readFrontmatter(lines []string) (frontmatterMetadata, int) {
	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if line == "---" {
			frontmatterLines := lines[1:index]

			frontmatter, err := parseYAMLFrontmatter(frontmatterLines)
			if err != nil {
				frontmatter = parseFrontmatterFallback(frontmatterLines)
			}

			return frontmatter, index + 1
		}
	}

	return parseFrontmatterFallback(lines[1:]), len(lines)
}

func parseYAMLFrontmatter(lines []string) (frontmatterMetadata, error) {
	if len(lines) == 0 {
		return frontmatterMetadata{}, nil
	}

	var rawFrontmatter map[string]any
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &rawFrontmatter); err != nil {
		return frontmatterMetadata{}, err
	}

	return frontmatterMetadata{
		Name:        scalarFrontmatterValue(rawFrontmatter["name"]),
		Description: scalarFrontmatterValue(rawFrontmatter["description"]),
		Tags:        tagsFrontmatterValue(rawFrontmatter["tags"]),
	}, nil
}

func parseFrontmatterFallback(lines []string) frontmatterMetadata {
	var frontmatter frontmatterMetadata

	for _, rawLine := range lines {
		key, value, ok := strings.Cut(strings.TrimSpace(rawLine), ":")
		if !ok {
			continue
		}

		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}

		normalizedValue := strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.ToLower(trimmedKey) {
		case "name":
			frontmatter.Name = normalizedValue
		case "description":
			frontmatter.Description = normalizedValue
		case "tags":
			frontmatter.Tags = splitTags(normalizedValue)
		}
	}

	return frontmatter
}

func scalarFrontmatterValue(value any) string {
	if value == nil {
		return ""
	}

	scalarValue := strings.TrimSpace(fmt.Sprint(value))
	if scalarValue == "<nil>" {
		return ""
	}

	return scalarValue
}

func tagsFrontmatterValue(value any) []string {
	switch typedValue := value.(type) {
	case nil:
		return nil
	case string:
		return splitTags(typedValue)
	case []any:
		tags := make([]string, 0, len(typedValue))
		for _, item := range typedValue {
			tags = append(tags, scalarFrontmatterValue(item))
		}
		return normalizeTags(tags...)
	case []string:
		return normalizeTags(typedValue...)
	default:
		return splitTags(scalarFrontmatterValue(typedValue))
	}
}
func splitTags(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tags = append(tags, strings.TrimSpace(part))
	}

	return normalizeTags(tags...)
}
