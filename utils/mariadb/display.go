package mariadb

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/utils/terminal"
)

// DisplayVersions displays MariaDB version information based on the configuration
func DisplayVersions(result *mariadb.VersionCheckResult, config *VersionConfig) error {
	switch config.OutputFormat {
	case "json":
		return displayVersionsJSON(result)
	case "simple":
		return displayVersionsSimple(result)
	default:
		return displayVersionsTable(result, config.ShowDetails)
	}
}

// displayVersionsJSON outputs version information in JSON format
func displayVersionsJSON(result *mariadb.VersionCheckResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// displayVersionsSimple outputs version information in simple text format
func displayVersionsSimple(result *mariadb.VersionCheckResult) error {
	terminal.PrintSuccess("✅ MariaDB Available Versions")
	fmt.Printf("\nCurrent Stable: %s\n", result.CurrentStable)
	fmt.Printf("Latest Version: %s\n", result.LatestVersion)
	fmt.Printf("Latest Minor: %s\n", result.LatestMinor)
	fmt.Printf("\nSupported Versions:\n")

	for _, version := range result.AvailableVersions {
		status := ""
		switch version.Type {
		case "stable":
			status = "📦"
		case "rolling":
			status = "🔄"
		case "rc":
			status = "🧪"
		}
		fmt.Printf("  %s %s (%s)\n", status, version.Version, version.Type)
	}

	fmt.Printf("\nChecked at: %s\n", result.CheckTime.Format("2006-01-02 15:04:05"))
	return nil
}

// displayVersionsTable outputs version information in table format
func displayVersionsTable(result *mariadb.VersionCheckResult, showDetails bool) error {
	terminal.ClearAndShowHeader("MariaDB Version Information")

	// Show summary first
	terminal.PrintSubHeader("📋 Summary")
	terminal.PrintInfo(fmt.Sprintf("Current Stable Version: %s%s%s",
		terminal.ColorGreen, result.CurrentStable, terminal.ColorReset))
	terminal.PrintInfo(fmt.Sprintf("Latest Available Version: %s%s%s",
		terminal.ColorBlue, result.LatestVersion, terminal.ColorReset))
	terminal.PrintInfo(fmt.Sprintf("Latest Minor Version: %s%s%s",
		terminal.ColorPurple, result.LatestMinor, terminal.ColorReset))
	terminal.PrintInfo(fmt.Sprintf("Total Versions Available: %s%d%s",
		terminal.ColorCyan, len(result.AvailableVersions), terminal.ColorReset))

	fmt.Println()

	// Show versions table
	terminal.PrintSubHeader("📦 Available Versions")

	headers := []string{"Version", "Type", "Status"}
	if showDetails {
		headers = append(headers, "Release Date")
	}

	var data [][]string
	for _, version := range result.AvailableVersions {
		status := "Available"
		if version.Version == result.CurrentStable {
			status = "Current Stable"
		} else if version.Version == result.LatestVersion {
			status = "Latest"
		} else if version.Version == result.LatestMinor {
			status = "Latest Minor"
		}

		row := []string{
			version.Version,
			getVersionTypeDisplay(version.Type),
			getStatusDisplay(status),
		}

		if showDetails {
			releaseDate := version.ReleaseDate
			if releaseDate == "" {
				releaseDate = "N/A"
			}
			row = append(row, releaseDate)
		}

		data = append(data, row)
	}

	terminal.FormatTable(headers, data)

	// Show additional information
	fmt.Println()
	terminal.PrintSubHeader("ℹ️  Version Types")
	terminal.PrintInfo("📦 Stable: Production-ready releases")
	terminal.PrintInfo("🔄 Rolling: Latest development version")
	terminal.PrintInfo("🧪 RC: Release candidate versions")

	fmt.Println()
	terminal.PrintInfo(fmt.Sprintf("Last checked: %s", result.CheckTime.Format("2006-01-02 15:04:05")))

	return nil
}

// getVersionTypeDisplay returns a formatted display string for version type
func getVersionTypeDisplay(versionType string) string {
	switch versionType {
	case "stable":
		return fmt.Sprintf("%s📦 Stable%s", terminal.ColorGreen, terminal.ColorReset)
	case "rolling":
		return fmt.Sprintf("%s🔄 Rolling%s", terminal.ColorBlue, terminal.ColorReset)
	case "rc":
		return fmt.Sprintf("%s🧪 RC%s", terminal.ColorYellow, terminal.ColorReset)
	default:
		return versionType
	}
}

// getStatusDisplay returns a formatted display string for version status
func getStatusDisplay(status string) string {
	switch status {
	case "Current Stable":
		return fmt.Sprintf("%s✅ %s%s", terminal.ColorGreen, status, terminal.ColorReset)
	case "Latest":
		return fmt.Sprintf("%s🆕 %s%s", terminal.ColorBlue, status, terminal.ColorReset)
	case "Latest Minor":
		return fmt.Sprintf("%s🔥 %s%s", terminal.ColorPurple, status, terminal.ColorReset)
	default:
		return fmt.Sprintf("%s⚪ %s%s", terminal.ColorWhite, status, terminal.ColorReset)
	}
}

// GenericVersionResult represents a generic version check result
type GenericVersionResult struct {
	Versions  map[string]VersionInfo `json:"versions"`
	Meta      GenericMetaInfo        `json:"meta"`
	FetchedAt time.Time              `json:"fetched_at"`
}

// GenericMetaInfo represents generic metadata
type GenericMetaInfo struct {
	OSInfo    OSInfo    `json:"os_info"`
	Sources   []string  `json:"sources"`
	Count     int       `json:"count"`
	FetchedAt time.Time `json:"fetched_at"`
}

// OSInfo represents OS information
type OSInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// DisplayVersionsFromGenericResult displays version check results based on configuration
func DisplayVersionsFromGenericResult(result *GenericVersionResult, config *VersionConfig) error {
	versions := result.Versions
	if len(versions) == 0 {
		terminal.PrintWarning("No MariaDB versions found")
		return nil
	}

	// Sort versions
	sortedVersions := make([]string, 0, len(versions))
	for version := range versions {
		sortedVersions = append(sortedVersions, version)
	}
	sort.Slice(sortedVersions, func(i, j int) bool {
		return CompareVersions(sortedVersions[i], sortedVersions[j])
	})

	if config.Format == "json" {
		return displayGenericVersionsJSON(versions, result.Meta)
	}

	return displayGenericVersionsTable(sortedVersions, versions, result.Meta, config)
}

func displayGenericVersionsJSON(versions map[string]VersionInfo, meta GenericMetaInfo) error {
	// Convert to JSON format
	output := map[string]interface{}{
		"versions": versions,
		"meta":     meta,
		"count":    len(versions),
	}

	if jsonBytes, err := json.MarshalIndent(output, "", "  "); err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	} else {
		fmt.Println(string(jsonBytes))
	}

	return nil
}

func displayGenericVersionsTable(sortedVersions []string, versions map[string]VersionInfo, meta GenericMetaInfo, config *VersionConfig) error {
	// Display header
	terminal.PrintSubHeader("📦 Available MariaDB Versions")
	fmt.Println(strings.Repeat("=", 50))

	// Display versions
	for i, version := range sortedVersions {
		info := versions[version]

		// Format version line
		line := fmt.Sprintf("%2d. MariaDB %s", i+1, version)

		// Add type information
		if info.Type != "" {
			line += fmt.Sprintf(" (%s)", info.Type)
		}

		fmt.Println(line)
	}

	// Display metadata
	fmt.Println()
	terminal.PrintSubHeader("ℹ️  Source Information")
	fmt.Printf("- Total versions found: %d\n", len(versions))
	fmt.Printf("- OS detected: %s\n", meta.OSInfo.OS)
	fmt.Printf("- Architecture: %s\n", meta.OSInfo.Arch)
	fmt.Printf("- Fetched at: %s\n", meta.FetchedAt.Format("2006-01-02 15:04:05 MST"))

	// Display sources used
	if len(meta.Sources) > 0 {
		fmt.Println("- Data sources:")
		for _, source := range meta.Sources {
			fmt.Printf("  • %s\n", source)
		}
	}

	return nil
}
