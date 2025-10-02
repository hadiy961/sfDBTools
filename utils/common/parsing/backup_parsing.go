// File : utils/common/parsing/backup_parsing.go
// Description : Parsing khusus untuk flags backup, termasuk opsi umum dan all-databases.
// Author : Hadiyatna Muflihun

package parsing

import (
	"fmt"
	defaultconfig "sfDBTools/internal/config/default_config"
	"sfDBTools/utils/common/structs"

	"github.com/spf13/cobra"
)

func ParseGeneralBackupFlags(cmd *cobra.Command) (generalDefaults *structs.BackupGeneralOptions, err error) {
	// 1. Dapatkan nilai default dari Configuration.
	// Nilai ini akan menjadi fallback terakhir.
	generalDefaults, err = defaultconfig.GetBackupGeneralDefaults()
	if err != nil {
		// Kembalikan error agar caller (seperti fungsi Run Cobra) yang menanganinya.
		return nil, fmt.Errorf("failed to load general backup defaults from config: %w", err)
	}

	// 2. Parse flags dinamis ke dalam struct menggunakan refleksi.
	// Ini akan menimpa nilai default dari config jika flag disediakan.
	// Pastikan generalDefaults sudah diinisialisasi (tidak nil) sebelum diteruskan.
	if err := DynamicParseFlags(cmd, generalDefaults); err != nil {
		return nil, fmt.Errorf("failed to dynamically parse general backup flags: %w", err)
	}

	return generalDefaults, nil
}

func ParseBackupAllDBFlags(cmd *cobra.Command) (opts *structs.BackupAllDBOptions, err error) {
	// 1. Dapatkan nilai default dari Configuration. Nilai ini menjadi baseline (fallback terendah).
	AllDBDefaults, err := defaultconfig.GetBackupAllDBDefaults()
	if err != nil {
		return nil, fmt.Errorf("gagal memuat default konfigurasi AllDB: %w", err)
	}
	opts = AllDBDefaults

	// 2. Dapatkan opsi umum (general) dan tetapkan ke dalam opts.
	generalOpts, err := ParseGeneralBackupFlags(cmd)
	if err != nil {
		return nil, err
	}
	opts.BackupGeneralOptions = *generalOpts

	// 3. Parse flags dinamis ke dalam struct menggunakan refleksi.
	if err := DynamicParseFlags(cmd, opts); err != nil {
		return nil, fmt.Errorf("gagal memproses flags AllDB secara dinamis: %w", err)
	}

	return opts, nil
}
