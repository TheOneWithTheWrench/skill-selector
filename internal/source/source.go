package source

import (
	"fmt"
	"strings"
)

// Source is a configured upstream skill source with a stable ID and fetch details.
type Source struct {
	id       string
	locator  string
	cloneURL string
	ref      string
	subpath  string
}

// Parse accepts supported source locators and returns a configured Source.
func Parse(rawURL string) (Source, error) {
	return parseGitHubTreeSource(rawURL)
}

// ID returns the stable source identifier used across persistence and sync.
func (s Source) ID() string {
	return s.id
}

// Locator returns the canonical source locator that was configured by the user.
func (s Source) Locator() string {
	return s.locator
}

// CloneURL returns the git remote used to materialize local mirrors for the source.
func (s Source) CloneURL() string {
	return s.cloneURL
}

// Ref returns the git ref that should be mirrored locally.
func (s Source) Ref() string {
	return s.ref
}

// Subpath returns the subtree inside the mirrored repository that should be scanned.
func (s Source) Subpath() string {
	return s.subpath
}

func newSource(id string, locator string, cloneURL string, ref string, subpath string) (Source, error) {
	if strings.TrimSpace(id) == "" {
		return Source{}, fmt.Errorf("source id required")
	}
	if strings.TrimSpace(locator) == "" {
		return Source{}, fmt.Errorf("source locator required")
	}
	if strings.TrimSpace(cloneURL) == "" {
		return Source{}, fmt.Errorf("source clone url required")
	}
	if strings.TrimSpace(ref) == "" {
		return Source{}, fmt.Errorf("source ref required")
	}

	return Source{
		id:       strings.TrimSpace(id),
		locator:  strings.TrimSpace(locator),
		cloneURL: strings.TrimSpace(cloneURL),
		ref:      strings.TrimSpace(ref),
		subpath:  strings.TrimSpace(subpath),
	}, nil
}

func sanitizeID(value string) string {
	var (
		builder     strings.Builder
		lastWasDash = true
	)

	for _, char := range strings.ToLower(value) {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
			lastWasDash = false
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
			lastWasDash = false
		default:
			if lastWasDash {
				continue
			}

			builder.WriteRune('-')
			lastWasDash = true
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "source"
	}

	return result
}
