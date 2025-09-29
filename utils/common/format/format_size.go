package format

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// ============================================================================
// SIZE FORMAT CONSTANTS
// ============================================================================

const (
	SizeDecimal = "decimal" // KB, MB, GB (1000-based)
	SizeBinary  = "binary"  // KiB, MiB, GiB (1024-based)
)

// ============================================================================
// FILE SIZE FORMATTING (using go-humanize)
// ============================================================================

// FormatSize formats bytes in human readable format based on the specified format.
//
// Supported formats:
//   - "decimal": KB, MB, GB (1000-based) - 1.0 MB
//   - "binary":  KiB, MiB, GiB (1024-based) - 1.0 MiB
//
// Example:
//
//	size := uint64(1024 * 1024 * 1024)
//	fmt.Println(FormatSize(size, SizeDecimal))  // 1.1 GB
//	fmt.Println(FormatSize(size, SizeBinary))   // 1.0 GiB
func FormatSize(bytes uint64, format string) string {
	switch format {
	case SizeBinary:
		return humanize.IBytes(bytes) // 1.0 GiB (1024-based)
	case SizeDecimal:
		return humanize.Bytes(bytes) // 1.0 GB (1000-based)
	default:
		return humanize.Bytes(bytes) // Default to decimal
	}
}

// FormatSizeRange formats a size range (e.g., "1.5 MB - 2.3 MB").
//
// Example:
//
//	min := uint64(1024 * 1024)
//	max := uint64(2 * 1024 * 1024)
//	fmt.Println(FormatSizeRange(min, max, SizeDecimal))  // 1.0 MB - 2.1 MB
func FormatSizeRange(minBytes, maxBytes uint64, format string) string {
	return fmt.Sprintf("%s - %s",
		FormatSize(minBytes, format),
		FormatSize(maxBytes, format))
}

// FormatTransferRate formats bytes per second as transfer rate.
//
// Example:
//
//	bytesPerSecond := uint64(1024 * 1024 * 10)
//	fmt.Println(FormatTransferRate(bytesPerSecond, SizeDecimal))  // 10.5 MB/s
func FormatTransferRate(bytesPerSecond uint64, format string) string {
	return FormatSize(bytesPerSecond, format) + "/s"
}

// ParseSize parses a human-readable size string to bytes.
// Supports formats like "1.5 MB", "2 GiB", "500 KB", etc.
//
// Example:
//
//	bytes, err := ParseSize("1.5 GB")
//	fmt.Println(bytes)  // 1500000000
func ParseSize(sizeStr string) (uint64, error) {
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}
	return bytes, nil
}

// FormatSizeWithPrecision formats bytes with a specific precision using
// binary units (KiB, MiB, ...). This mirrors the behavior of
// format.FormatSizeWithPrecision which accepts an int64 and returns
// values like "1.23 MiB".

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
