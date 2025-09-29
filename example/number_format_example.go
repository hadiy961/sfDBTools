package example

import (
	"fmt"
	"sfDBTools/utils/common/format"
)

func Number_format() {
	// ===== NUMBER FORMATTING =====
	fmt.Println("Number Formatting:")
	fmt.Println(format.FormatNumber(1234567))                    // 1,234,567
	fmt.Println(format.FormatNumberWithLocale(1234567, "id-ID")) // 1.234.567
	fmt.Println(format.FormatOrdinal(21))                        // 21st

	// ===== PERCENTAGE =====
	fmt.Println("\nPercentage:")
	fmt.Println(format.FormatPercent(75.5))              // 75.50%
	fmt.Println(format.FormatPercentFromRatio(0.755, 1)) // 75.5%
	fmt.Println(format.FormatPercentChange(15.5))        // +15.50%
	fmt.Println(format.FormatPercentChange(-8.3))        // -8.30%

	// ===== CURRENCY =====
	fmt.Println("\nCurrency:")
	fmt.Println(format.FormatCurrency(1234567.89, "USD"))                    // $1,234,567.89
	fmt.Println(format.FormatCurrency(1234567.89, "IDR"))                    // Rp1,234,567.89
	fmt.Println(format.FormatCurrencyWithLocale(1234567.89, "IDR", "id-ID")) // Rp1.234.567,89
	fmt.Println(format.FormatCurrencyShort(1234567, "USD"))                  // $1.23M

	// ===== FILE SIZE =====
	fmt.Println("\nFile Size:")
	fmt.Println(format.FormatBytes(1024 * 1024 * 1024)) // 1.0 GB
	fmt.Println(format.FormatBytesIEC(1024 * 1024))     // 1.0 MiB

	// ===== LARGE NUMBERS =====
	fmt.Println("\nLarge Numbers:")
	fmt.Println(format.FormatNumberShort(1234567))    // 1.23M
	fmt.Println(format.FormatNumberShort(1234567890)) // 1.23B
	fmt.Println(format.FormatNumberSI(1234567))       // 1.235M

	// ===== FRACTION & RATIO =====
	fmt.Println("\nFraction & Ratio:")
	fmt.Println(format.FormatFraction(3, 4)) // 3/4
	fmt.Println(format.FormatRatio(16, 9))   // 16:9

	// ===== ROUNDING =====
	fmt.Println("\nRounding:")
	fmt.Println(format.RoundToDecimal(3.14159, 2)) // 3.14
	fmt.Println(format.RoundToNearest(127, 10))    // 130

	// ===== SCIENTIFIC =====
	fmt.Println("\nScientific:")
	fmt.Println(format.FormatScientific(1234567, 2)) // 1.23e+06
}
