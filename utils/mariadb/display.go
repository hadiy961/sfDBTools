package mariadb

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/utils/terminal"
)

func DisplayVersions(result *mariadb.VersionCheckResult, config *VersionConfig) error {
	switch config.OutputFormat {
	case "json":
		return displayJSON(result)
	case "simple":
		return displaySimple(result)
	default:
		return displayTable(result, config.ShowDetails)
	}
}

func displayJSON(result *mariadb.VersionCheckResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func displaySimple(result *mariadb.VersionCheckResult) error {
	terminal.PrintSuccess("✅ MariaDB Available Versions")
	fmt.Printf("\nCurrent Stable: %s\n", result.CurrentStable)
	fmt.Printf("Latest Version: %s\n", result.LatestVersion)
	fmt.Printf("Latest Minor: %s\n", result.LatestMinor)
	fmt.Printf("\nSupported Versions:\n")

	for _, version := range result.AvailableVersions {
		status := getVersionIcon(version.Type)
		fmt.Printf("  %s %s (%s)\n", status, version.Version, version.Type)
	}

	fmt.Printf("\nChecked at: %s\n", result.CheckTime.Format("2006-01-02 15:04:05"))
	return nil
}

func displayTable(result *mariadb.VersionCheckResult, showDetails bool) error {
	terminal.ClearAndShowHeader("MariaDB Version Information")

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
	terminal.PrintSubHeader("📦 Available Versions")

	headers := []string{"Version", "Type", "Status"}
	if showDetails {
		headers = append(headers, "Release Date")
	}

	var data [][]string
	for _, version := range result.AvailableVersions {
		status := getVersionStatus(version.Version, result)
		row := []string{
			version.Version,
			formatVersionType(version.Type),
			formatStatus(status),
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
	printLegend()
	terminal.PrintInfo(fmt.Sprintf("Last checked: %s", result.CheckTime.Format("2006-01-02 15:04:05")))
	return nil
}

func DisplayVersionsFromGenericResult(result *GenericVersionResult, config *VersionConfig) error {
	if len(result.Versions) == 0 {
		terminal.PrintWarning("No MariaDB versions found")
		return nil
	}

	sortedVersions := make([]string, 0, len(result.Versions))
	for version := range result.Versions {
		sortedVersions = append(sortedVersions, version)
	}
	sort.Slice(sortedVersions, func(i, j int) bool {
		return CompareVersions(sortedVersions[i], sortedVersions[j])
	})

	if config.Format == "json" {
		return displayGenericJSON(result.Versions, result.Meta)
	}
	return displayGenericTable(sortedVersions, result.Versions, result.Meta, config)
}

func displayGenericJSON(versions map[string]VersionInfo, meta GenericMetaInfo) error {
	output := map[string]interface{}{
		"versions": versions,
		"meta":     meta,
		"count":    len(versions),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func displayGenericTable(sortedVersions []string, versions map[string]VersionInfo, meta GenericMetaInfo, config *VersionConfig) error {
	terminal.ClearAndShowHeader("MariaDB Version Information")

	terminal.PrintSubHeader("📋 Summary")
	terminal.PrintInfo(fmt.Sprintf("Total Versions Available: %s%d%s",
		terminal.ColorCyan, len(versions), terminal.ColorReset))
	terminal.PrintInfo(fmt.Sprintf("OS detected: %s%s%s",
		terminal.ColorGreen, meta.OSInfo.OS, terminal.ColorReset))
	terminal.PrintInfo(fmt.Sprintf("Architecture: %s%s%s",
		terminal.ColorBlue, meta.OSInfo.Arch, terminal.ColorReset))

	fmt.Println()
	terminal.PrintSubHeader("📦 Available Versions")

	headers := []string{"Version", "Type", "Release Date"}
	if config.ShowDetails {
		headers = append(headers, "EOL Date", "Status")
	}

	var data [][]string
	for _, version := range sortedVersions {
		info := versions[version]
		row := []string{
			version,
			formatVersionType(info.Type),
			getDateOrNA(info.ReleaseDate),
		}

		if config.ShowDetails {
			row = append(row, getDateOrNA(info.EOLDate), formatEOLStatus(info.EOLDate))
		}
		data = append(data, row)
	}

	terminal.FormatTable(headers, data)
	printMetaInfo(meta)
	printLegend()
	if config.ShowDetails {
		printEOLLegend()
	}
	return nil
}

func getVersionIcon(versionType string) string {
	switch versionType {
	case "stable":
		return "📦"
	case "rolling":
		return "🔄"
	case "rc":
		return "🧪"
	default:
		return "⚪"
	}
}

func getVersionStatus(version string, result *mariadb.VersionCheckResult) string {
	if version == result.CurrentStable {
		return "Current Stable"
	}
	if version == result.LatestVersion {
		return "Latest"
	}
	if version == result.LatestMinor {
		return "Latest Minor"
	}
	return "Available"
}

func formatVersionType(versionType string) string {
	icon := getVersionIcon(versionType)
	color := getTypeColor(versionType)
	return fmt.Sprintf("%s%s %s%s", color, icon, versionType, terminal.ColorReset)
}

func formatStatus(status string) string {
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

func formatEOLStatus(eolDate string) string {
	if eolDate == "No LTS" {
		return fmt.Sprintf("%s🔄 No LTS%s", terminal.ColorBlue, terminal.ColorReset)
	}
	if eolDate == "TBD" || eolDate == "N/A" || eolDate == "" {
		return fmt.Sprintf("%s❓ N/A%s", terminal.ColorWhite, terminal.ColorReset)
	}

	eolTime, err := time.Parse("2006-01-02", eolDate)
	if err != nil {
		return fmt.Sprintf("%s❓ Invalid%s", terminal.ColorWhite, terminal.ColorReset)
	}

	now := time.Now()
	if eolTime.Before(now) {
		return fmt.Sprintf("%s❌ EOL%s", terminal.ColorRed, terminal.ColorReset)
	}

	if eolTime.Before(now.AddDate(0, 6, 0)) {
		return fmt.Sprintf("%s⚠️ EOL Soon%s", terminal.ColorYellow, terminal.ColorReset)
	}

	return fmt.Sprintf("%s✅ Supported%s", terminal.ColorGreen, terminal.ColorReset)
}

func getTypeColor(versionType string) string {
	switch versionType {
	case "stable":
		return terminal.ColorGreen
	case "rolling":
		return terminal.ColorBlue
	case "rc":
		return terminal.ColorYellow
	default:
		return terminal.ColorWhite
	}
}

func getDateOrNA(date string) string {
	if date == "" {
		return "N/A"
	}
	return date
}

func printLegend() {
	fmt.Println()
	terminal.PrintSubHeader("ℹ️ Version Types")
	terminal.PrintInfo("📦 Stable: Production-ready releases")
	terminal.PrintInfo("🔄 Rolling: Latest development version")
	terminal.PrintInfo("🧪 RC: Release candidate versions")
	fmt.Println()
}

func printEOLLegend() {
	fmt.Println()
	terminal.PrintSubHeader("🕐 Support Status")
	terminal.PrintInfo("✅ Supported: Version is currently supported")
	terminal.PrintInfo("⚠️ EOL Soon: Support ends within 6 months")
	terminal.PrintInfo("❌ EOL: Version is no longer supported")
	terminal.PrintInfo("🔄 No LTS: Rolling/RC versions have no long-term support")
}

func printMetaInfo(meta GenericMetaInfo) {
	fmt.Println()
	terminal.PrintSubHeader("ℹ️ Source Information")
	terminal.PrintInfo(fmt.Sprintf("Fetched at: %s", meta.FetchedAt.Format("2006-01-02 15:04:05 MST")))

	if len(meta.Sources) > 0 {
		terminal.PrintInfo("Data sources:")
		for _, source := range meta.Sources {
			terminal.PrintInfo(fmt.Sprintf("  • %s", source))
		}
	}
}

type GenericVersionResult struct {
	Versions  map[string]VersionInfo `json:"versions"`
	Meta      GenericMetaInfo        `json:"meta"`
	FetchedAt time.Time              `json:"fetched_at"`
}

type GenericMetaInfo struct {
	OSInfo    OSInfo    `json:"os_info"`
	Sources   []string  `json:"sources"`
	Count     int       `json:"count"`
	FetchedAt time.Time `json:"fetched_at"`
}

type OSInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}
