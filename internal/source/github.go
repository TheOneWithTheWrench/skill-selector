package source

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	neturl "net/url"
	"path"
	"strings"
)

func parseGitHubSource(raw string) (Source, error) {
	normalizedURL := strings.TrimSpace(raw)
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
	if len(segments) < 2 {
		return Source{}, fmt.Errorf("source url must be a GitHub repo or tree url: %q", normalizedURL)
	}

	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	if owner == "" || repo == "" {
		return Source{}, fmt.Errorf("source url is missing owner or repo: %q", normalizedURL)
	}

	ref := ""
	subpath := ""
	switch {
	case len(segments) == 2:
		// GitHub repo root. We intentionally mirror the default branch by cloning
		// without an explicit branch and scanning the repository root.
	case len(segments) >= 4 && segments[2] == "tree":
		ref = segments[3]
		if ref == "" {
			return Source{}, fmt.Errorf("source url is missing ref: %q", normalizedURL)
		}
		if len(segments) > 4 {
			subpath = path.Clean(strings.Join(segments[4:], "/"))
			if subpath == "." {
				subpath = ""
			}
		}
	default:
		return Source{}, fmt.Errorf("source url must be a GitHub repo or tree url: %q", normalizedURL)
	}

	canonicalLocator := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	if ref != "" {
		canonicalLocator += "/tree/" + ref
		if subpath != "" {
			canonicalLocator += "/" + subpath
		}
	}

	sourceID := newGitSourceID(owner, repo, ref, subpath)
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	return newSource(sourceID, canonicalLocator, cloneURL, ref, subpath)
}

func newGitSourceID(owner string, repo string, ref string, subpath string) string {
	var base strings.Builder
	base.WriteString(owner)
	base.WriteString("-")
	base.WriteString(repo)

	if subpath != "" {
		base.WriteString("-")
		base.WriteString(path.Base(subpath))
	}

	hash := sha1.Sum([]byte(owner + "/" + repo + "@" + ref + ":" + subpath))
	return sanitizeID(base.String()) + "-" + hex.EncodeToString(hash[:4])
}
