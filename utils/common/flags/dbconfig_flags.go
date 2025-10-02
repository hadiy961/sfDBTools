package flags

import (
	"fmt"
	"os"
	defaultconfig "sfDBTools/internal/config/default_config"

	"github.com/spf13/cobra"
)

// AddGenerateFlags adds flags specific to the generate command
func AddGenerateFlags(cmd *cobra.Command) {
	// 1. Dapatkan default options dari konfigurasi
	defaultOpt, err := defaultconfig.GetDBConfigGenerateDefaults()
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

// AddCommonDbConfigFlags adds shared flags used across dbconfig commands
func AddCommonDbConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Specific encrypted config file")
}

// AddDeleteFlags adds flags specific to the delete command
func AddDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip confirmation prompts")
	cmd.Flags().Bool("all", false, "Delete all encrypted config files")
}
