// file utils/common/format/format_time.go
// Utility functions for formatting time using mature libraries
package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
	"github.com/jinzhu/now"
)

// ============================================================================
// DATE/TIME FORMAT CONSTANTS
// ============================================================================

const (
	// Date formats
	DateISO      = "2006-01-02"
	DateUS       = "01/02/2006"
	DateEU       = "02/01/2006"
	DateReadable = "January 2, 2006"
	DateShort    = "Jan 2, 2006"
	DateCompact  = "20060102"

	// Time formats
	Time24      = "15:04:05"
	Time12      = "3:04:05 PM"
	TimeShort24 = "15:04"
	TimeShort12 = "3:04 PM"

	// DateTime combined
	DateTime     = "2006-01-02 15:04:05"
	DateTimeISO  = "2006-01-02T15:04:05Z07:00"
	DateTimeFull = "Monday, January 2, 2006 at 3:04:05 PM"

	// Special formats
	UnixTimestamp = "unix"
	RFC3339Format = "rfc3339"
)

// ============================================================================
// DURATION FORMATTING (using durafmt)
// ============================================================================

// FormatDuration formats a duration using durafmt library.
//
// Options:
//   - "default": "2 hours 30 minutes 45 seconds"
//   - "short": "2h30m45s"
//   - "long": Full verbose format
//
// Example:
//
//	d := 2*time.Hour + 30*time.Minute + 45*time.Second
//	fmt.Println(FormatDuration(d, "default")) // 2 hours 30 minutes 45 seconds
//	fmt.Println(FormatDuration(d, "short"))   // 2h30m45s
func FormatDuration(d time.Duration, format string) string {
	switch format {
	case "compact":
		return strings.ReplaceAll(d.Round(time.Second).String(), " ", "")
	case "hms":
		return FormatHMS(d, false)
	case "hms-ms":
		return FormatHMS(d, true)
	case "short":
		return d.String() // Go's native: 2h30m45s
	case "long":
		return durafmt.Parse(d).String() // Full: 2 hours 30 minutes 45 seconds
	case "details":
		return FormatDurationPrecise(d, 4) // Full: 2 hours 30 minutes 45 seconds
	case "limit":
		// Limit to first 2 units
		return durafmt.Parse(d).LimitFirstN(2).String() // 2 hours 30 minutes
	default:
		return durafmt.Parse(d).String()
	}
}

func FormatHMS(d time.Duration, withMillis bool) string {
	h, m, s := int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60
	if withMillis {
		ms := d.Milliseconds() % 1000
		return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// FormatDurationPrecise formats duration with custom precision.
//
// Example:
//
//	d := 2*time.Hour + 30*time.Minute + 45*time.Second + 123*time.Millisecond
//	fmt.Println(FormatDurationPrecise(d, 2)) // 2 hours 30 minutes
//	fmt.Println(FormatDurationPrecise(d, 4)) // 2 hours 30 minutes 45 seconds 123 milliseconds
func FormatDurationPrecise(d time.Duration, precision int) string {
	return durafmt.Parse(d).LimitFirstN(precision).String()
}

// FormatElapsedTime formats elapsed time from a start time.
//
// Example:
//
//	start := time.Now().Add(-2 * time.Hour)
//	fmt.Println(FormatElapsedTime(start)) // 2 hours 0 minutes
func FormatElapsedTime(startTime time.Time) string {
	return durafmt.Parse(time.Since(startTime)).LimitFirstN(2).String()
}

// FormatRemainingTime formats remaining time until a deadline.
//
// Example:
//
//	deadline := time.Now().Add(30 * time.Minute)
//	fmt.Println(FormatRemainingTime(deadline)) // 30 minutes 0 seconds
func FormatRemainingTime(deadline time.Time) string {
	remaining := time.Until(deadline)
	if remaining < 0 {
		return "overdue by " + durafmt.Parse(-remaining).LimitFirstN(2).String()
	}
	return durafmt.Parse(remaining).LimitFirstN(2).String()
}

// ============================================================================
// TIMESTAMP FORMATTING (using jinzhu/now)
// ============================================================================

// FormatTime formats a time.Time into a string based on the specified format.
//
// Example:
//
//	t := time.Now()
//	fmt.Println(FormatTime(t, DateISO))       // 2024-03-15
//	fmt.Println(FormatTime(t, DateTime))      // 2024-03-15 14:30:45
//	fmt.Println(FormatTime(t, UnixTimestamp)) // 1710512445
func FormatTime(t time.Time, format string) string {
	switch format {
	case UnixTimestamp:
		return fmt.Sprintf("%d", t.Unix())
	case RFC3339Format:
		return t.Format(time.RFC3339)
	default:
		return t.Format(format)
	}
}

// FormatRelativeTime formats time relative to now using go-humanize.
// Returns strings like "2 hours ago", "in 3 days", "just now".
//
// Example:
//
//	past := time.Now().Add(-2 * time.Hour)
//	fmt.Println(FormatRelativeTime(past))    // "2 hours ago"
//	future := time.Now().Add(3 * time.Hour)
//	fmt.Println(FormatRelativeTime(future))  // "3 hours from now"
func FormatRelativeTime(t time.Time) string {
	return humanize.Time(t)
}

// FormatRelativeTimeShort formats time relative to now in short form.
// Returns strings like "2h ago", "3d from now".
//
// Example:
//
//	past := time.Now().Add(-2 * time.Hour)
//	fmt.Println(FormatRelativeTimeShort(past)) // "2 hours ago"
func FormatRelativeTimeShort(t time.Time) string {
	// go-humanize doesn't have short format, use custom
	now := time.Now()
	diff := now.Sub(t)

	if diff >= 0 {
		return durafmt.Parse(diff).LimitFirstN(1).String() + " ago"
	}
	return "in " + durafmt.Parse(-diff).LimitFirstN(1).String()
}

// FormatSmartDate returns "Today", "Yesterday", "Tomorrow", or formatted date.
//
// Example:
//
//	today := time.Now()
//	fmt.Println(FormatSmartDate(today))      // "Today"
//	yesterday := today.AddDate(0, 0, -1)
//	fmt.Println(FormatSmartDate(yesterday))  // "Yesterday"
func FormatSmartDate(t time.Time) string {
	today := now.BeginningOfDay()
	inputDay := now.New(t).BeginningOfDay()

	switch {
	case inputDay.Equal(today):
		return "Today"
	case inputDay.Equal(today.AddDate(0, 0, -1)):
		return "Yesterday"
	case inputDay.Equal(today.AddDate(0, 0, 1)):
		return "Tomorrow"
	case inputDay.After(now.BeginningOfWeek()) && inputDay.Before(now.EndOfWeek()):
		return t.Format("Monday") // Day name for this week
	case t.Year() == time.Now().Year():
		return t.Format("Jan 2")
	default:
		return t.Format("Jan 2, 2006")
	}
}

// FormatCalendar formats time in calendar style with time.
// e.g., "Today at 2:30 PM", "Yesterday at 10:15 AM", "Mar 15 at 3:45 PM"
//
// Example:
//
//	t := time.Now().Add(-2 * time.Hour)
//	fmt.Println(FormatCalendar(t)) // "Today at 12:30 PM"
func FormatCalendar(t time.Time) string {
	dateStr := FormatSmartDate(t)
	timeStr := t.Format("3:04 PM")
	return fmt.Sprintf("%s at %s", dateStr, timeStr)
}

// FormatTimeRange formats a time range intelligently using jinzhu/now.
//
// Example:
//
//	start := time.Now()
//	end := start.Add(2 * time.Hour)
//	fmt.Println(FormatTimeRange(start, end))
//	// "Mar 15, 2024 2:30 PM - 4:30 PM" (same day)
func FormatTimeRange(start, end time.Time) string {
	startDay := now.New(start).BeginningOfDay()
	endDay := now.New(end).BeginningOfDay()

	if startDay.Equal(endDay) {
		return fmt.Sprintf("%s %s - %s",
			start.Format("Jan 2, 2006"),
			start.Format("3:04 PM"),
			end.Format("3:04 PM"))
	}

	return fmt.Sprintf("%s - %s",
		start.Format("Jan 2, 2006 3:04 PM"),
		end.Format("Jan 2, 2006 3:04 PM"))
}

// ============================================================================
// TIMEZONE UTILITIES
// ============================================================================

// FormatTimeWithZone formats time with timezone information.
//
// Supported formats:
//   - "full":   2024-03-15 14:30:45 WIB
//   - "short":  14:30 WIB
//   - "offset": 14:30:45 +0700
//
// Example:
//
//	t := time.Now()
//	fmt.Println(FormatTimeWithZone(t, "full")) // "2024-03-15 14:30:45 WIB"
func FormatTimeWithZone(t time.Time, format string) string {
	switch format {
	case "full":
		return t.Format("2006-01-02 15:04:05 MST")
	case "short":
		return t.Format("15:04 MST")
	case "offset":
		return t.Format("15:04:05 -0700")
	default:
		return t.Format(format)
	}
}

// ConvertTimezone converts time to specific timezone.
//
// Common timezones: "Asia/Jakarta", "America/New_York", "Europe/London", "UTC"
//
// Example:
//
//	t := time.Now()
//	jakarta, _ := ConvertTimezone(t, "Asia/Jakarta")
//	fmt.Println(jakarta)
func ConvertTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, fmt.Errorf("invalid timezone: %w", err)
	}
	return t.In(loc), nil
}

// ============================================================================
// TIME RANGE UTILITIES (using jinzhu/now)
// ============================================================================

// BeginningOfDay returns the beginning of the day (00:00:00).
func BeginningOfDay(t time.Time) time.Time {
	return now.New(t).BeginningOfDay()
}

// EndOfDay returns the end of the day (23:59:59).
func EndOfDay(t time.Time) time.Time {
	return now.New(t).EndOfDay()
}

// BeginningOfWeek returns the beginning of the week (Sunday 00:00:00).
func BeginningOfWeek(t time.Time) time.Time {
	return now.New(t).BeginningOfWeek()
}

// EndOfWeek returns the end of the week (Saturday 23:59:59).
func EndOfWeek(t time.Time) time.Time {
	return now.New(t).EndOfWeek()
}

// BeginningOfMonth returns the beginning of the month.
func BeginningOfMonth(t time.Time) time.Time {
	return now.New(t).BeginningOfMonth()
}

// EndOfMonth returns the end of the month.
func EndOfMonth(t time.Time) time.Time {
	return now.New(t).EndOfMonth()
}

// BeginningOfYear returns the beginning of the year.
func BeginningOfYear(t time.Time) time.Time {
	return now.New(t).BeginningOfYear()
}

// EndOfYear returns the end of the year.
func EndOfYear(t time.Time) time.Time {
	return now.New(t).EndOfYear()
}
