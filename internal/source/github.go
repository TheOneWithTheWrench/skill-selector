package source

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	neturl "net/url"
	"path"
	"strings"
)

func parseGitHubTreeSource(raw string) (Source, error) {
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
	if len(segments) < 4 {
		return Source{}, fmt.Errorf("source url must be a GitHub tree url: %q", normalizedURL)
	}

	if segments[2] != "tree" {
		return Source{}, fmt.Errorf("source url must contain /tree/: %q", normalizedURL)
	}

	owner := segments[0]
	repo := segments[1]
	ref := segments[3]
	if owner == "" || repo == "" || ref == "" {
		return Source{}, fmt.Errorf("source url is missing owner, repo, or ref: %q", normalizedURL)
	}

	subpath := ""
	if len(segments) > 4 {
		subpath = path.Clean(strings.Join(segments[4:], "/"))
		if subpath == "." {
			subpath = ""
		}
	}

	sourceID := newGitSourceID(owner, repo, ref, subpath)
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	return newSource(sourceID, normalizedURL, cloneURL, ref, subpath)
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
