package common

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// Size formatting
func FormatSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}

func FormatSizeWithPrecision(bytes int64, precision int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div, exp = div*unit, exp+1
	}

	return fmt.Sprintf("%.*f %ciB", precision, float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatSpeed(bytesPerSecond float64) string {
	return FormatSize(int64(bytesPerSecond)) + "/s"
}

// Duration formatting
func FormatDuration(d time.Duration, format string) string {
	switch format {
	case "compact":
		return strings.ReplaceAll(d.Round(time.Second).String(), " ", "")
	case "hms":
		return formatHMS(d, false)
	case "hms-ms":
		return formatHMS(d, true)
	case "words":
		return formatDurationWords(d, false)
	case "words-ms":
		return formatDurationWords(d, true)
	default:
		return d.Round(time.Second).String()
	}
}

func formatHMS(d time.Duration, withMillis bool) string {
	h, m, s := int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60
	if withMillis {
		ms := d.Milliseconds() % 1000
		return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func formatDurationWords(d time.Duration, withMillis bool) string {
	var parts []string

	if h := int(d.Hours()); h > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", h, pluralize("hour", h)))
	}
	if m := int(d.Minutes()) % 60; m > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", m, pluralize("minute", m)))
	}

	s := int(d.Seconds()) % 60
	if withMillis {
		if ms := d.Milliseconds() % 1000; ms > 0 && s == 0 && len(parts) == 0 {
			parts = append(parts, fmt.Sprintf("%d %s", ms, pluralize("millisecond", int(ms))))
		}
	}

	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d %s", s, pluralize("second", s)))
	}

	return strings.Join(parts, " ")
}

// Number formatting
func FormatPercent(value float64, precision ...int) string {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}
	return fmt.Sprintf("%.*f%%", p, value)
}

func FormatNumber(number interface{}, precision ...int) string {
	switch v := number.(type) {
	case int64:
		return humanize.Comma(v)
	case float64:
		p := 2
		if len(precision) > 0 {
			p = precision[0]
		}
		formatted := fmt.Sprintf("%.*f", p, v)
		if parts := strings.Split(formatted, "."); len(parts) > 1 {
			intPart, _ := strconv.ParseInt(parts[0], 10, 64)
			return humanize.Comma(intPart) + "." + parts[1]
		}
		intPart, _ := strconv.ParseInt(formatted, 10, 64)
		return humanize.Comma(intPart)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Time formatting
func FormatTime(t time.Time, format string) string {
	switch format {
	case "iso":
		return t.Format("2006-01-02")
	case "iso-time":
		return t.Format("2006-01-02 15:04:05")
	case "iso-tz":
		return t.Format("2006-01-02T15:04:05-07:00")
	case "relative":
		return humanize.Time(t)
	default:
		return t.Format(format)
	}
}

// Utility formatting
func FormatProgressBar(progress float64, width int) string {
	progress = math.Max(0, math.Min(1, progress))
	completed := int(math.Round(float64(width) * progress))
	bar := strings.Repeat("█", completed) + strings.Repeat("░", width-completed)
	return fmt.Sprintf("[%s] %s", bar, FormatPercent(progress*100))
}

func FormatOrdinal(n int) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}

	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}

	return fmt.Sprintf("%d%s", n, suffix)
}

func FormatBool(value bool, trueStr, falseStr string) string {
	if value {
		return trueStr
	}
	return falseStr
}

// Helper functions
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// parseMemorySizeToMB mengkonversi string memory size ke MB.
// Mendukung format seperti: "128M", "1536MB", "1.5G", "2GB", "1024K", "1T", "512".
func ParseMemorySizeToMB(size string) (int, error) {
	if strings.TrimSpace(size) == "" {
		return 0, fmt.Errorf("empty memory size")
	}

	s := strings.ToUpper(strings.TrimSpace(size))
	// remove trailing B if present (e.g., MB, GB)
	if strings.HasSuffix(s, "B") && len(s) > 1 {
		s = strings.TrimSuffix(s, "B")
	}

	// find suffix (last alpha run)
	lastAlpha := -1
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			lastAlpha = i
		} else {
			break
		}
	}

	var numStr string
	var suffix string
	if lastAlpha == -1 {
		// no suffix; treat as MB if reasonable (or bytes? choose MB for safety)
		numStr = s
		suffix = "M"
	} else {
		numStr = strings.TrimSpace(s[:lastAlpha])
		suffix = strings.TrimSpace(s[lastAlpha:])
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid memory size: missing numeric part in %q", size)
	}

	// parse float to accept "1.5"
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric part in memory size %q: %w", numStr, err)
	}

	// convert to MB
	var mb float64
	switch suffix {
	case "K", "KB":
		mb = num / 1024.0
	case "M", "MB":
		mb = num
	case "G", "GB":
		mb = num * 1024.0
	case "T", "TB":
		mb = num * 1024.0 * 1024.0
	default:
		return 0, fmt.Errorf("unsupported memory size suffix: %s", suffix)
	}

	// round to nearest integer MB
	result := int(math.Round(mb))
	if result < 0 {
		return 0, fmt.Errorf("computed negative memory size from %q", size)
	}
	return result, nil
}
