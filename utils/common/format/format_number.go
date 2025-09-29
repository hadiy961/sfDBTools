// file utils/common/format/format_number.go
// Utility functions for formatting numbers using mature libraries
package format

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
	"github.com/leekchan/accounting"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ============================================================================
// NUMBER FORMATTING (using go-humanize)
// ============================================================================

// FormatNumber formats a number with thousand separators.
//
// Example:
//
//	fmt.Println(FormatNumber(1234567))     // 1,234,567
//	fmt.Println(FormatNumber(1234567.89))  // 1,234,567.89
func FormatNumber(n interface{}) string {
	switch v := n.(type) {
	case int:
		return humanize.Comma(int64(v))
	case int64:
		return humanize.Comma(v)
	case float64:
		return humanize.Commaf(v)
	case float32:
		return humanize.Commaf(float64(v))
	default:
		return fmt.Sprintf("%v", n)
	}
}

// FormatNumberWithLocale formats a number with locale-specific thousand separators.
//
// Supported locales: "en-US", "id-ID", "de-DE", "fr-FR", etc.
//
// Example:
//
//	fmt.Println(FormatNumberWithLocale(1234567, "en-US"))  // 1,234,567
//	fmt.Println(FormatNumberWithLocale(1234567, "id-ID"))  // 1.234.567
//	fmt.Println(FormatNumberWithLocale(1234567, "de-DE"))  // 1.234.567
func FormatNumberWithLocale(n interface{}, locale string) string {
	tag := language.Make(locale)
	p := message.NewPrinter(tag)

	switch v := n.(type) {
	case int:
		return p.Sprintf("%d", v)
	case int64:
		return p.Sprintf("%d", v)
	case float64:
		return p.Sprintf("%.2f", v)
	case float32:
		return p.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", n)
	}
}

// FormatOrdinal formats a number as an ordinal (1st, 2nd, 3rd, etc.).
//
// Example:
//
//	fmt.Println(FormatOrdinal(1))   // 1st
//	fmt.Println(FormatOrdinal(2))   // 2nd
//	fmt.Println(FormatOrdinal(3))   // 3rd
//	fmt.Println(FormatOrdinal(21))  // 21st
func FormatOrdinal(n int) string {
	return humanize.Ordinal(n)
}

// ============================================================================
// PERCENTAGE FORMATTING
// ============================================================================

// FormatPercent formats a number as a percentage with custom precision.
//
// Example:
//
//	fmt.Println(FormatPercent(75.5))           // 75.50%
//	fmt.Println(FormatPercent(75.5678, 1))     // 75.6%
//	fmt.Println(FormatPercent(75.5678, 3))     // 75.568%
func FormatPercent(value float64, precision ...int) string {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}
	return fmt.Sprintf("%.*f%%", p, value)
}

// FormatPercentFromRatio formats a ratio (0.0 - 1.0) as percentage.
//
// Example:
//
//	fmt.Println(FormatPercentFromRatio(0.755))      // 75.50%
//	fmt.Println(FormatPercentFromRatio(0.755, 1))   // 75.5%
//	fmt.Println(FormatPercentFromRatio(1.0))        // 100.00%
func FormatPercentFromRatio(ratio float64, precision ...int) string {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}
	return fmt.Sprintf("%.*f%%", p, ratio*100)
}

// FormatPercentChange formats percentage change (with +/- sign).
//
// Example:
//
//	fmt.Println(FormatPercentChange(15.5))    // +15.50%
//	fmt.Println(FormatPercentChange(-8.3))    // -8.30%
//	fmt.Println(FormatPercentChange(0))       // 0.00%
func FormatPercentChange(change float64, precision ...int) string {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}

	if change > 0 {
		return fmt.Sprintf("+%.*f%%", p, change)
	}
	return fmt.Sprintf("%.*f%%", p, change)
}

// ============================================================================
// CURRENCY FORMATTING (using leekchan/accounting)
// ============================================================================

// FormatCurrency formats a number as currency with symbol.
//
// Supported currencies: "USD", "EUR", "IDR", "JPY", etc.
//
// Example:
//
//	fmt.Println(FormatCurrency(1234567.89, "USD"))  // $1,234,567.89
//	fmt.Println(FormatCurrency(1234567.89, "IDR"))  // Rp1,234,567.89
//	fmt.Println(FormatCurrency(1234567.89, "EUR"))  // €1,234,567.89
func FormatCurrency(amount float64, currency string) string {
	ac := getCurrencyFormatter(currency)
	return ac.FormatMoney(amount)
}

// FormatCurrencyWithLocale formats currency with locale-specific formatting.
//
// Example:
//
//	fmt.Println(FormatCurrencyWithLocale(1234567.89, "USD", "en-US"))  // $1,234,567.89
//	fmt.Println(FormatCurrencyWithLocale(1234567.89, "IDR", "id-ID"))  // Rp1.234.567,89
//	fmt.Println(FormatCurrencyWithLocale(1234567.89, "EUR", "de-DE"))  // 1.234.567,89 €
func FormatCurrencyWithLocale(amount float64, currency, locale string) string {
	ac := getCurrencyFormatterWithLocale(currency, locale)
	return ac.FormatMoney(amount)
}

// getCurrencyFormatter returns accounting formatter for specific currency
func getCurrencyFormatter(currency string) accounting.Accounting {
	switch currency {
	case "USD":
		return accounting.Accounting{Symbol: "$", Precision: 2}
	case "EUR":
		return accounting.Accounting{Symbol: "€", Precision: 2}
	case "GBP":
		return accounting.Accounting{Symbol: "£", Precision: 2}
	case "JPY":
		return accounting.Accounting{Symbol: "¥", Precision: 0}
	case "IDR":
		return accounting.Accounting{Symbol: "Rp", Precision: 2}
	case "CNY":
		return accounting.Accounting{Symbol: "¥", Precision: 2}
	case "INR":
		return accounting.Accounting{Symbol: "₹", Precision: 2}
	default:
		return accounting.Accounting{Symbol: currency + " ", Precision: 2}
	}
}

// getCurrencyFormatterWithLocale returns formatter with locale-specific settings
func getCurrencyFormatterWithLocale(currency, locale string) accounting.Accounting {
	ac := getCurrencyFormatter(currency)

	// Adjust thousand/decimal separator based on locale
	switch locale {
	case "id-ID", "de-DE", "es-ES", "fr-FR":
		// Use . for thousands and , for decimals
		ac.Thousand = "."
		ac.Decimal = ","
	case "en-US", "en-GB":
		// Use , for thousands and . for decimals
		ac.Thousand = ","
		ac.Decimal = "."
	default:
		ac.Thousand = ","
		ac.Decimal = "."
	}

	return ac
}

// FormatCurrencyShort formats currency in short form (K, M, B).
//
// Example:
//
//	fmt.Println(FormatCurrencyShort(1234, "USD"))       // $1.23K
//	fmt.Println(FormatCurrencyShort(1234567, "USD"))    // $1.23M
//	fmt.Println(FormatCurrencyShort(1234567890, "USD")) // $1.23B
func FormatCurrencyShort(amount float64, currency string) string {
	symbol := getCurrencySymbol(currency)
	shortened := FormatNumberShort(amount)
	return symbol + shortened
}

func getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥",
		"IDR": "Rp", "CNY": "¥", "INR": "₹",
	}
	if symbol, ok := symbols[currency]; ok {
		return symbol
	}
	return currency + " "
}

// ============================================================================
// FILE SIZE FORMATTING (using go-humanize)
// ============================================================================

// FormatBytes formats bytes in human readable format (KB, MB, GB, etc.).
//
// Example:
//
//	fmt.Println(FormatBytes(1024))              // 1.0 kB
//	fmt.Println(FormatBytes(1024 * 1024))       // 1.0 MB
//	fmt.Println(FormatBytes(1024 * 1024 * 1024))// 1.0 GB
func FormatBytes(bytes uint64) string {
	return humanize.Bytes(bytes)
}

// FormatBytesIEC formats bytes using IEC standard (KiB, MiB, GiB).
//
// Example:
//
//	fmt.Println(FormatBytesIEC(1024))           // 1.0 KiB
//	fmt.Println(FormatBytesIEC(1024 * 1024))    // 1.0 MiB
func FormatBytesIEC(bytes uint64) string {
	return humanize.IBytes(bytes)
}

// ============================================================================
// LARGE NUMBER FORMATTING
// ============================================================================

// FormatNumberShort formats large numbers in short form (K, M, B, T).
//
// Example:
//
//	fmt.Println(FormatNumberShort(1234))        // 1.23K
//	fmt.Println(FormatNumberShort(1234567))     // 1.23M
//	fmt.Println(FormatNumberShort(1234567890))  // 1.23B
func FormatNumberShort(n float64) string {
	abs := math.Abs(n)
	sign := ""
	if n < 0 {
		sign = "-"
	}

	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%s%.2fT", sign, abs/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%s%.2fB", sign, abs/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%s%.2fM", sign, abs/1e6)
	case abs >= 1e3:
		return fmt.Sprintf("%s%.2fK", sign, abs/1e3)
	default:
		return fmt.Sprintf("%s%.0f", sign, abs)
	}
}

// FormatNumberSI formats numbers using SI prefixes (k, M, G, T).
//
// Example:
//
//	fmt.Println(FormatNumberSI(1234))        // 1.234k
//	fmt.Println(FormatNumberSI(1234567))     // 1.235M
func FormatNumberSI(n float64) string {
	return humanize.SIWithDigits(n, 3, "")
}

// ============================================================================
// FRACTION & RATIO FORMATTING
// ============================================================================

// FormatFraction formats a fraction (e.g., 3/4, 1/2).
//
// Example:
//
//	fmt.Println(FormatFraction(3, 4))  // 3/4
//	fmt.Println(FormatFraction(1, 2))  // 1/2
func FormatFraction(numerator, denominator int) string {
	return fmt.Sprintf("%d/%d", numerator, denominator)
}

// FormatRatio formats a ratio (e.g., 16:9, 4:3).
//
// Example:
//
//	fmt.Println(FormatRatio(16, 9))  // 16:9
//	fmt.Println(FormatRatio(4, 3))   // 4:3
func FormatRatio(a, b int) string {
	gcd := greatestCommonDivisor(a, b)
	return fmt.Sprintf("%d:%d", a/gcd, b/gcd)
}

func greatestCommonDivisor(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// ============================================================================
// ROUNDING UTILITIES
// ============================================================================

// RoundToDecimal rounds a float to specified decimal places.
//
// Example:
//
//	fmt.Println(RoundToDecimal(3.14159, 2))  // 3.14
//	fmt.Println(RoundToDecimal(3.14159, 4))  // 3.1416
func RoundToDecimal(value float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	return math.Round(value*shift) / shift
}

// RoundToNearest rounds to the nearest multiple.
//
// Example:
//
//	fmt.Println(RoundToNearest(123, 10))    // 120
//	fmt.Println(RoundToNearest(127, 10))    // 130
//	fmt.Println(RoundToNearest(1234, 100))  // 1200
func RoundToNearest(value, multiple float64) float64 {
	return math.Round(value/multiple) * multiple
}

// ============================================================================
// SCIENTIFIC NOTATION
// ============================================================================

// FormatScientific formats a number in scientific notation.
//
// Example:
//
//	fmt.Println(FormatScientific(1234567, 2))   // 1.23e+06
//	fmt.Println(FormatScientific(0.000123, 2))  // 1.23e-04
func FormatScientific(value float64, precision int) string {
	return fmt.Sprintf("%.*e", precision, value)
}
