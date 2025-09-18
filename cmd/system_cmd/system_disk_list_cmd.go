package system_cmd

import (
	"fmt"

	"sfDBTools/utils/common"
	"sfDBTools/utils/disk"

	"github.com/spf13/cobra"
)

var SystemDiskListCmd = &cobra.Command{
	Use:   "disk-list",
	Short: "List semua mount/partisi dan penggunaan (mirip df -h)",
	Run: func(cmd *cobra.Command, args []string) {
		partitions, err := disk.GetAllPartitions()
		if err != nil {
			fmt.Printf("Failed to list disk usages: %v\n", err)
			return
		}
		fmt.Printf("%-30s %-8s %-8s %-8s %-6s %-6s\n", "Filesystem", "Size", "Used", "Avail", "Use%", "Type")
		for _, u := range partitions {
			fmt.Printf("%-30s %-8s %-8s %-8s %-6.1f %-6s\n",
				u.Mountpoint,
				common.FormatSizeWithPrecision(u.Total, 1),
				common.FormatSizeWithPrecision(u.Used, 1),
				common.FormatSizeWithPrecision(u.Free, 1),
				u.UsedPercent,
				u.Fstype,
			)
		}
	},
}

func init() {
	// registered by parent
}
