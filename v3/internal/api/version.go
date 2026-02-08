package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// GitHubReleasesURL is the API endpoint for latest release
	githubReleasesURL = "https://api.github.com/repos/shinokada/tera/releases/latest"
	// ReleasePageURL is the URL users can visit to download
	ReleasePageURL = "https://github.com/shinokada/tera/releases/latest"
)

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
}

// VersionChecker checks for new versions on GitHub
type VersionChecker struct {
	httpClient *http.Client
}

// NewVersionChecker creates a new version checker
func NewVersionChecker() *VersionChecker {
	return &VersionChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestRelease fetches the latest release info from GitHub
func (vc *VersionChecker) GetLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubReleasesURL, nil)
	if err != nil {
		return nil, err
	}

	// GitHub API recommends setting User-Agent
	req.Header.Set("User-Agent", "tera-radio-player")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := vc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// CompareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Normalize versions (remove 'v' prefix if present)
	v1 = normalizeVersion(v1)
	v2 = normalizeVersion(v2)

	// Handle dev versions
	if v1 == "dev" {
		return -1 // dev is always older
	}
	if v2 == "dev" {
		return 1 // any version is newer than dev
	}

	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		p1, p2 := 0, 0
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}
		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	// Handle pre-release versions (e.g., v1.0.0-rc.1)
	// A version without pre-release is greater than one with pre-release
	hasPreRelease1 := strings.Contains(v1, "-")
	hasPreRelease2 := strings.Contains(v2, "-")

	if hasPreRelease1 && !hasPreRelease2 {
		return -1 // v1.0.0-rc.1 < v1.0.0
	}
	if !hasPreRelease1 && hasPreRelease2 {
		return 1 // v1.0.0 > v1.0.0-rc.1
	}

	return 0
}

// normalizeVersion removes the 'v' prefix and trims whitespace
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

// parseVersion parses a version string into numeric parts
func parseVersion(v string) []int {
	// Remove pre-release suffix for base version comparison
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	result := make([]int, 0, 3)

	for _, p := range parts {
		var num int
		if _, err := fmt.Sscanf(p, "%d", &num); err == nil {
			result = append(result, num)
		}
	}

	return result
}

// IsNewerVersion checks if latestVersion is newer than currentVersion
func IsNewerVersion(currentVersion, latestVersion string) bool {
	return CompareVersions(currentVersion, latestVersion) < 0
}
