package catalog

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
)

// Scan walks a source mirror and discovers skill directories backed by `SKILL.md` files.
func Scan(mirror source.Mirror) (Skills, error) {
	rootPath := mirror.SubtreePath()

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("stat source subtree %q: %w", rootPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source subtree is not a directory: %q", rootPath)
	}

	var discoveredSkills Skills
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

		name, description, err := readSkillMetadata(skillFilePath, filepath.Base(currentPath))
		if err != nil {
			return err
		}

		identity, err := skill_identity.New(mirror.ID(), relativePath)
		if err != nil {
			return err
		}

		discoveredSkill, err := NewSkill(identity, name, description)
		if err != nil {
			return err
		}

		discoveredSkills = append(discoveredSkills, discoveredSkill)

		return filepath.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("scan source %q: %w", mirror.ID(), err)
	}

	return NewSkills(discoveredSkills...), nil
}

func shouldSkipDirectory(name string, currentPath string, rootPath string) bool {
	if currentPath == rootPath {
		return false
	}

	if name == ".git" {
		return true
	}

	return strings.HasPrefix(name, ".")
}

func readSkillMetadata(path string, fallbackName string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("open skill file %q: %w", path, err)
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
		return "", "", fmt.Errorf("scan skill file %q: %w", path, err)
	}

	startIndex := 0
	nameFromHeading := false
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		frontmatter, nextIndex := readFrontmatter(lines)
		startIndex = nextIndex

		if frontmatterName := strings.TrimSpace(frontmatter["name"]); frontmatterName != "" {
			name = frontmatterName
		}
		if frontmatterDescription := strings.TrimSpace(frontmatter["description"]); frontmatterDescription != "" {
			description = frontmatterDescription
		}
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
			if description != "" {
				break
			}
			continue
		}

		if description == "" && !strings.HasPrefix(line, "#") {
			description = line
			break
		}
	}

	return name, description, nil
}

func readFrontmatter(lines []string) (map[string]string, int) {
	frontmatter := make(map[string]string)

	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if line == "---" {
			return frontmatter, index + 1
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}

		frontmatter[trimmedKey] = normalizeFrontmatterValue(strings.TrimSpace(value))
	}

	return frontmatter, len(lines)
}

func normalizeFrontmatterValue(value string) string {
	if value == "" {
		return value
	}

	if unquotedValue, err := strconv.Unquote(value); err == nil {
		return unquotedValue
	}

	return strings.Trim(value, `"'`)
}
