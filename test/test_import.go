package main

import (
	"fmt"
	"sfDBTools/utils/common"
)

func main() {
	size := int64(1024)
	formatted := common.FormatSize(size)
	fmt.Printf("Formatted size: %s\n", formatted)
}
