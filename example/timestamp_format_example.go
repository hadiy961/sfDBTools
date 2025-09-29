package example

import (
	"fmt"
	"sfDBTools/utils/common/format"
	"time"
)

func Timestamp_format() {
	now := time.Now()

	// ===== DURATION =====
	d := 2*time.Hour + 30*time.Minute + 45*time.Second
	fmt.Println("Duration:")
	fmt.Println(format.FormatDuration(d, "short"))  // 2h30m45s
	fmt.Println(format.FormatDuration(d, "long"))   // 2 hours 30 minutes 45 seconds
	fmt.Println(format.FormatDurationPrecise(d, 2)) // 2 hours 30 minutes

	// ===== RELATIVE TIME =====
	past := now.Add(-2 * time.Hour)
	future := now.Add(3 * time.Hour)
	fmt.Println("\nRelative Time:")
	fmt.Println(format.FormatRelativeTime(past))   // 2 hours ago
	fmt.Println(format.FormatRelativeTime(future)) // 3 hours from now

	// ===== SMART DATE =====
	fmt.Println("\nSmart Date:")
	fmt.Println(format.FormatSmartDate(now))                   // Today
	fmt.Println(format.FormatSmartDate(now.AddDate(0, 0, -1))) // Yesterday
	fmt.Println(format.FormatCalendar(now))                    // Today at 2:30 PM

	// ===== TIME RANGE =====
	fmt.Println("\nTime Range Utilities:")
	fmt.Println(format.BeginningOfDay(now))   // 2024-03-15 00:00:00
	fmt.Println(format.EndOfDay(now))         // 2024-03-15 23:59:59
	fmt.Println(format.BeginningOfWeek(now))  // Sunday 00:00:00
	fmt.Println(format.BeginningOfMonth(now)) // 2024-03-01 00:00:00

	// ===== HUMANIZE =====
	fmt.Println("\nHumanize:")
	fmt.Println(format.FormatBytes(1024 * 1024)) // 1.0 MB
	fmt.Println(format.FormatNumber(1234567))    // 1,234,567
	fmt.Println(format.FormatOrdinal(21))        // 21st
}
