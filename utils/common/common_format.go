package common

import (
	"fmt"
	"math"
	"sfDBTools/utils/common/format"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

// Size formatting
func FormatSize(bytes int64) string {
	format := format.FormatSize(uint64(bytes), format.SizeBinary)
	return format
}

func FormatSpeed(bytesPerSecond float64) string {
	format := format.FormatTransferRate(uint64(bytesPerSecond), format.SizeBinary)
	return format
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

func ParseSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if len(sizeStr) == 0 {
		return 0
	}

	// Extract number dan suffix
	var number int64
	var suffix string

	for i, char := range sizeStr {
		if char < '0' || char > '9' {
			var err error
			number, err = strconv.ParseInt(sizeStr[:i], 10, 64)
			if err != nil {
				return 0
			}
			suffix = sizeStr[i:]
			break
		}
	}

	// Jika tidak ada suffix, anggap sudah dalam bytes
	if suffix == "" {
		var err error
		number, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0
		}
		return number
	}

	// Convert berdasarkan suffix
	switch suffix {
	case "K", "KB":
		return number * 1024
	case "M", "MB":
		return number * 1024 * 1024
	case "G", "GB":
		return number * 1024 * 1024 * 1024
	case "T", "TB":
		return number * 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
}
