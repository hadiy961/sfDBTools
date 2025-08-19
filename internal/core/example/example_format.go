package example

import (
	"fmt"
	"sfDBTools/utils/common"
	"time"
)

// DemoAllFormattingFeatures mendemonstrasikan semua fitur formatting yang tersedia
func DemoAllFormattingFeatures() error {
	fmt.Println("=== Demo Format Examples ===")

	// Demo size formatting
	DemoSizeFormatting()

	// Demo duration formatting
	DemoDurationFormatting()

	// Demo number formatting
	DemoNumberFormatting()

	// Demo time formatting
	DemoTimeFormatting()

	// Demo utility formatting
	DemoUtilityFormatting()

	fmt.Println("=== Demo Format Examples Complete ===")
	return nil
}

// DemoSizeFormatting mendemonstrasikan formatting untuk ukuran file
func DemoSizeFormatting() {
	fmt.Println("--- Size Formatting Demo ---")

	sizes := []int64{1024, 1048576, 1073741824, 5497558138880} // 1KB, 1MB, 1GB, 5TB

	for _, size := range sizes {
		fmt.Printf("Size %d bytes:\n", size)
		fmt.Printf("  - Standard: %s\n", common.FormatSize(size))
		fmt.Printf("  - Precision 1: %s\n", common.FormatSizeWithPrecision(size, 1))
		fmt.Printf("  - Precision 3: %s\n", common.FormatSizeWithPrecision(size, 3))
	}

	// Demo speed formatting
	fmt.Println("Speed Formatting:")
	speeds := []float64{1024.5, 1048576.7, 52428800.0} // ~1KB/s, ~1MB/s, ~50MB/s
	for _, speed := range speeds {
		fmt.Printf("  - Speed %.1f bytes/s: %s\n", speed, common.FormatSpeed(speed))
	}
}

// DemoDurationFormatting mendemonstrasikan formatting untuk durasi waktu
func DemoDurationFormatting() {
	fmt.Println("--- Duration Formatting Demo ---")

	durations := []time.Duration{
		45 * time.Second,
		2*time.Minute + 30*time.Second,
		1*time.Hour + 23*time.Minute + 45*time.Second,
		350 * time.Millisecond,
	}

	formats := []string{"compact", "hms", "hms-ms", "words", "words-ms", "default"}

	for _, duration := range durations {
		fmt.Printf("Duration %v:\n", duration)
		for _, format := range formats {
			formatted := common.FormatDuration(duration, format)
			fmt.Printf("  - %s: %s\n", format, formatted)
		}
		fmt.Println("")
	}
}

// DemoNumberFormatting mendemonstrasikan formatting untuk angka dan persentase
func DemoNumberFormatting() {
	fmt.Println("--- Number Formatting Demo ---")

	// Demo percent formatting
	fmt.Println("Percent Formatting:")
	percentages := []float64{0.1234, 45.6789, 99.99, 100.0}
	for _, pct := range percentages {
		fmt.Printf("  - %.4f = %s (default)\n", pct, common.FormatPercent(pct))
		fmt.Printf("  - %.4f = %s (0 precision)\n", pct, common.FormatPercent(pct, 0))
		fmt.Printf("  - %.4f = %s (4 precision)\n", pct, common.FormatPercent(pct, 4))
	}

	// Demo number formatting
	fmt.Println("Number Formatting:")
	intNumbers := []int64{1234, 1234567, 1234567890}
	for _, num := range intNumbers {
		fmt.Printf("  - Integer %d: %s\n", num, common.FormatNumber(num))
	}

	floatNumbers := []float64{1234.56, 1234567.89, 1234567890.123}
	for _, num := range floatNumbers {
		fmt.Printf("  - Float %.3f: %s (default)\n", num, common.FormatNumber(num))
		fmt.Printf("  - Float %.3f: %s (0 precision)\n", num, common.FormatNumber(num, 0))
		fmt.Printf("  - Float %.3f: %s (4 precision)\n", num, common.FormatNumber(num, 4))
	}
}

// DemoTimeFormatting mendemonstrasikan formatting untuk waktu
func DemoTimeFormatting() {
	fmt.Println("--- Time Formatting Demo ---")

	now := time.Now()
	pastTime := now.Add(-2 * time.Hour)

	times := []time.Time{now, pastTime}
	formats := []string{"iso", "iso-time", "iso-tz", "relative", "2006-01-02 15:04:05"}

	for i, t := range times {
		label := "Now"
		if i == 1 {
			label = "2 hours ago"
		}
		fmt.Printf("Time (%s):\n", label)
		for _, format := range formats {
			formatted := common.FormatTime(t, format)
			fmt.Printf("  - %s: %s\n", format, formatted)
		}
		fmt.Println("")
	}
}

// DemoUtilityFormatting mendemonstrasikan utility formatting lainnya
func DemoUtilityFormatting() {
	fmt.Println("--- Utility Formatting Demo ---")

	// Demo progress bar
	fmt.Println("Progress Bar Formatting:")
	progresses := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, progress := range progresses {
		bar := common.FormatProgressBar(progress, 20)
		fmt.Printf("  - Progress %.0f%%: %s\n", progress*100, bar)
	}

	// Demo ordinal numbers
	fmt.Println("Ordinal Number Formatting:")
	numbers := []int{1, 2, 3, 4, 11, 12, 13, 21, 22, 23, 101, 102, 103}
	for _, num := range numbers {
		fmt.Printf("  - %d: %s\n", num, common.FormatOrdinal(num))
	}

	// Demo boolean formatting
	fmt.Println("Boolean Formatting:")
	bools := []bool{true, false}
	for _, b := range bools {
		fmt.Printf("  - %t: %s\n", b, common.FormatBool(b, "Yes", "No"))
		fmt.Printf("  - %t: %s\n", b, common.FormatBool(b, "Enabled", "Disabled"))
		fmt.Printf("  - %t: %s\n", b, common.FormatBool(b, "✓", "✗"))
	}
}

// DemoBackupScenarioFormat mendemonstrasikan penggunaan formatting dalam konteks backup
func DemoBackupScenarioFormat() {
	fmt.Println("--- Backup Scenario Demo ---")

	// Simulate backup statistics
	startTime := time.Now().Add(-15 * time.Minute)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	totalFiles := int64(1247)
	totalSize := int64(5497558138880) // ~5TB
	speed := float64(totalSize) / duration.Seconds()
	successRate := 98.7

	fmt.Println("Backup Statistics:")
	fmt.Printf("  - Duration: %s\n", common.FormatDuration(duration, "words"))
	fmt.Printf("  - Total Files: %s\n", common.FormatNumber(totalFiles))
	fmt.Printf("  - Total Size: %s\n", common.FormatSize(totalSize))
	fmt.Printf("  - Average Speed: %s\n", common.FormatSpeed(speed))
	fmt.Printf("  - Success Rate: %s\n", common.FormatPercent(successRate, 1))
	fmt.Printf("  - Started: %s\n", common.FormatTime(startTime, "iso-time"))
	fmt.Printf("  - Completed: %s\n", common.FormatTime(endTime, "relative"))

	// Progress simulation
	fmt.Println("Backup Progress:")
	for i := 0; i <= 10; i++ {
		progress := float64(i) / 10.0
		bar := common.FormatProgressBar(progress, 30)
		fmt.Printf("  Step %s: %s\n", common.FormatOrdinal(i+1), bar)
	}
}
