package mariadb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// SimpleVersionInfo represents a MariaDB version
type SimpleVersionInfo struct {
	Version     string `json:"version"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	SupportType string `json:"support_type"`
	EOLDate     string `json:"eol_date,omitempty"`
}

// MariaDBAPIResponse represents the response from MariaDB REST API
type MariaDBAPIResponse struct {
	MajorReleases []struct {
		ReleaseID          string `json:"release_id"`
		ReleaseName        string `json:"release_name"`
		ReleaseStatus      string `json:"release_status"`
		ReleaseSupportType string `json:"release_support_type"`
		ReleaseEOLDate     string `json:"release_eol_date"`
	} `json:"major_releases"`
}

// MariaDB API constants (using utils/mariadb/constants.go pattern)
const (
	MariaDBAPIURL    = "https://downloads.mariadb.org/rest-api/mariadb/"
	DefaultTimeout   = 30 * time.Second
	DefaultUserAgent = "sfDBTools/1.0 MariaDB-Version-Checker"
)

// GetAvailableVersions fetches available MariaDB versions from official REST API
func GetAvailableVersions() ([]SimpleVersionInfo, error) {
	lg, _ := logger.Get()
	lg.Info("Fetching MariaDB versions from official REST API")

	// Check internet connectivity first using utils
	if err := common.RequireInternetForOperation("MariaDB version check"); err != nil {
		return nil, fmt.Errorf("connectivity check failed: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	// Create request with context and user agent
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", MariaDBAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse JSON response
	var apiResponse MariaDBAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Convert to our format
	var versions []SimpleVersionInfo
	for _, release := range apiResponse.MajorReleases {
		version := SimpleVersionInfo{
			Version:     release.ReleaseID,
			Type:        determineVersionType(release.ReleaseStatus),
			Status:      release.ReleaseStatus,
			SupportType: release.ReleaseSupportType,
			EOLDate:     release.ReleaseEOLDate,
		}
		versions = append(versions, version)
	}

	// Sort versions (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[j].Version, versions[i].Version)
	})

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	lg.Info("Successfully fetched MariaDB versions", logger.Int("count", len(versions)))
	return versions, nil
}

// determineVersionType determines the type based on status
func determineVersionType(status string) string {
	switch strings.ToLower(status) {
	case "stable":
		return "stable"
	case "rc":
		return "rc"
	default:
		return "stable"
	}
}

// compareVersions compares two version strings (returns true if v1 < v2)
func compareVersions(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 < p2 {
			return true
		} else if p1 > p2 {
			return false
		}
	}
	return false
}

// DisplayVersions displays the versions in a simple format
func DisplayVersions(versions []SimpleVersionInfo, outputFormat string) error {
	switch outputFormat {
	case "json":
		return displayJSON(versions)
	case "simple":
		return displaySimple(versions)
	default:
		return displayTable(versions)
	}
}

func displayJSON(versions []SimpleVersionInfo) error {
	jsonData, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func displaySimple(versions []SimpleVersionInfo) error {
	fmt.Println("Available MariaDB versions:")
	for _, version := range versions {
		fmt.Printf("  - %s (%s)\n", version.Version, version.Status)
	}
	return nil
}

func displayTable(versions []SimpleVersionInfo) error {
	terminal.ClearAndShowHeader("MariaDB Available Versions")

	fmt.Printf("%-10s %-10s %-20s %-15s\n", "Version", "Status", "Support Type", "EOL Date")
	fmt.Println(strings.Repeat("-", 65))

	for _, version := range versions {
		eolDate := version.EOLDate
		if eolDate == "" {
			eolDate = "N/A"
		}

		statusColor := terminal.ColorGreen
		if version.Status == "RC" {
			statusColor = terminal.ColorYellow
		}

		fmt.Printf("%-10s %s%-10s%s %-20s %-15s\n",
			version.Version,
			statusColor,
			version.Status,
			terminal.ColorReset,
			version.SupportType,
			eolDate,
		)
	}

	fmt.Printf("\nTotal versions: %d\n", len(versions))
	fmt.Printf("Fetched at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}
