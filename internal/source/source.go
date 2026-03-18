package source

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	neturl "net/url"
	"path"
	"strings"
)

// Source is a configured upstream GitHub tree that can provide skills.
type Source struct {
	url     string
	owner   string
	repo    string
	ref     string
	subpath string
}

// Parse validates a GitHub tree URL and returns it as a configured Source.
func Parse(rawURL string) (Source, error) {
	normalizedURL := strings.TrimSpace(rawURL)
	if normalizedURL == "" {
		return Source{}, fmt.Errorf("source url required")
	}

	parsedURL, err := neturl.Parse(normalizedURL)
	if err != nil {
		return Source{}, fmt.Errorf("parse source url %q: %w", normalizedURL, err)
	}

	if parsedURL.Scheme != "https" {
		return Source{}, fmt.Errorf("source url must use https: %q", normalizedURL)
	}

	if parsedURL.Host != "github.com" {
		return Source{}, fmt.Errorf("source url must point at github.com: %q", normalizedURL)
	}

	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(segments) < 4 {
		return Source{}, fmt.Errorf("source url must be a GitHub tree url: %q", normalizedURL)
	}

	if segments[2] != "tree" {
		return Source{}, fmt.Errorf("source url must contain /tree/: %q", normalizedURL)
	}

	if segments[0] == "" || segments[1] == "" || segments[3] == "" {
		return Source{}, fmt.Errorf("source url is missing owner, repo, or ref: %q", normalizedURL)
	}

	subpath := ""
	if len(segments) > 4 {
		subpath = path.Clean(strings.Join(segments[4:], "/"))
		if subpath == "." {
			subpath = ""
		}
	}

	return Source{
		url:     normalizedURL,
		owner:   segments[0],
		repo:    segments[1],
		ref:     segments[3],
		subpath: subpath,
	}, nil
}

// URL returns the canonical URL string that is persisted for the source.
func (s Source) URL() string {
	return s.url
}

// Owner returns the GitHub owner for the source repository.
func (s Source) Owner() string {
	return s.owner
}

// Repo returns the GitHub repository name for the source.
func (s Source) Repo() string {
	return s.repo
}

// Ref returns the git ref selected by the source URL.
func (s Source) Ref() string {
	return s.ref
}

// Subpath returns the subtree inside the repository that should be scanned.
func (s Source) Subpath() string {
	return s.subpath
}

// RepoSlug returns the repository in owner/repo form for clone and display flows.
func (s Source) RepoSlug() string {
	return s.owner + "/" + s.repo
}

// ID derives a stable identifier from repo, ref, and subtree for deduplication and storage.
func (s Source) ID() string {
	var base strings.Builder
	base.WriteString(s.owner)
	base.WriteString("-")
	base.WriteString(s.repo)

	if s.subpath != "" {
		base.WriteString("-")
		base.WriteString(path.Base(s.subpath))
	}

	hash := sha1.Sum([]byte(s.RepoSlug() + "@" + s.ref + ":" + s.subpath))
	return sanitizeID(base.String()) + "-" + hex.EncodeToString(hash[:4])
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
