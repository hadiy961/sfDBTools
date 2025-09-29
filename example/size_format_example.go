package example

import (
	"fmt"
	"sfDBTools/utils/common/format"
)

func main() {
	size := uint64(1536 * 1024 * 1024) // 1.5 GiB

	// ===== SIZE FORMATTING =====
	fmt.Println("Size Formatting:")
	fmt.Println(format.FormatSize(size, format.SizeDecimal)) // 1.6 GB
	fmt.Println(format.FormatSize(size, format.SizeBinary))  // 1.5 GiB

	// ===== VARIOUS SIZES =====
	fmt.Println("\nVarious Sizes:")
	fmt.Println(format.FormatSize(1024, format.SizeBinary))                // 1.0 KiB
	fmt.Println(format.FormatSize(1024*1024, format.SizeBinary))           // 1.0 MiB
	fmt.Println(format.FormatSize(1024*1024*1024, format.SizeBinary))      // 1.0 GiB
	fmt.Println(format.FormatSize(1024*1024*1024*1024, format.SizeBinary)) // 1.0 TiB

	// Small sizes
	fmt.Println(format.FormatSize(512, format.SizeDecimal))  // 512 B
	fmt.Println(format.FormatSize(1023, format.SizeDecimal)) // 1.0 kB

	// ===== SIZE RANGE =====
	min := uint64(1024 * 1024)
	max := uint64(2 * 1024 * 1024)
	fmt.Println("\nSize Range:")
	fmt.Println(format.FormatSizeRange(min, max, format.SizeDecimal)) // 1.0 MB - 2.1 MB

	// ===== TRANSFER RATE =====
	bytesPerSecond := uint64(1024 * 1024 * 10)
	fmt.Println("\nTransfer Rate:")
	fmt.Println(format.FormatTransferRate(bytesPerSecond, format.SizeDecimal)) // 10.5 MB/s
	fmt.Println(format.FormatTransferRate(bytesPerSecond, format.SizeBinary))  // 10.0 MiB/s

	// ===== PARSE SIZE (BONUS!) =====
	fmt.Println("\nParse Size:")
	bytes1, _ := format.ParseSize("1.5 GB")
	fmt.Println(bytes1) // 1500000000

	bytes2, _ := format.ParseSize("2 GiB")
	fmt.Println(bytes2) // 2147483648

	bytes3, _ := format.ParseSize("500 MB")
	fmt.Println(bytes3) // 500000000
}
