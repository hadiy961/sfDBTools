// File : utils/common/flags/backup_flags.go
// Description : Definisi flags untuk operasi backup, termasuk opsi umum dan all-databases.
// Author : Hadiyatna Muflihun

package flags

import (
	"fmt"
	"os"
	defaultconfig "sfDBTools/internal/config/default_config"
	"sfDBTools/utils/common/structs"

	"github.com/spf13/cobra"
)

// AddGeneralBackupFlags menambahkan flags umum untuk semua jenis operasi backup.
func AddGeneralBackupFlags(cmd *cobra.Command, defaultOpt structs.BackupGeneralOptions) {
	if err := DynamicAddFlags(cmd, &defaultOpt); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering general flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddBackupAllDBFlags menambahkan flags spesifik untuk backup seluruh database (AllDB).
func AddBackupAllDBFlags(cmd *cobra.Command) {
	// 1. Dapatkan default options dari konfigurasi
	defaultOpt, err := defaultconfig.GetBackupAllDBDefaults()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get default all DB backup configuration:", err)
		os.Exit(1)
	}

	// 2. Daftarkan flags umum terlebih dahulu
	if err := DynamicAddFlags(cmd, defaultOpt); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering AllDB flags dynamically: %v\n", err)
		os.Exit(1)
	}

}
