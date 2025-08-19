package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	RestoreUtils "sfDBTools/internal/core/restore/utils"

	"github.com/spf13/cobra"
)

// ParseRestoreOptionsFromFlags parses restore options from command flags.
// Deprecated: Use ResolveRestoreConfig from utils/restore package instead
func ParseRestoreOptionsFromFlags(cmd *cobra.Command) (RestoreUtils.RestoreOptions, error) {
	target_host, target_port, target_user, target_password,
		_, _, _, _,
		_, _, _, _, _ := config.GetBackupDefaults()

	dbName := GetStringFlagOrEnv(cmd, "target_db", "TARGET_DB", "")
	if dbName == "" {
		return RestoreUtils.RestoreOptions{}, fmt.Errorf("database name is required")
	}

	host := GetStringFlagOrEnv(cmd, "target_host", "TARGET_HOST", target_host)
	port := GetIntFlagOrEnv(cmd, "target_port", "TARGET_PORT", target_port)
	user := GetStringFlagOrEnv(cmd, "target_user", "TARGET_USER", target_user)
	password := GetStringFlagOrEnv(cmd, "target_password", "TARGET_PASSWORD", target_password)
	file := GetStringFlagOrEnv(cmd, "file", "FILE", "")
	VerifyChecksum := GetBoolFlagOrEnv(cmd, "verify-checksum", "VERIFY_CHECKSUM", false)

	return RestoreUtils.RestoreOptions{
		Host:           host,
		Port:           port,
		User:           user,
		Password:       password,
		DBName:         dbName,
		File:           file,
		VerifyChecksum: VerifyChecksum,
	}, nil
}

func GetStringFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal string) string {
	val, _ := cmd.Flags().GetString(flagName)
	if val != "" {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		return env
	}
	return defaultVal
}

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal int) int {
	val, _ := cmd.Flags().GetInt(flagName)
	if val != 0 {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		// ignore error, fallback ke default jika gagal
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return defaultVal
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal bool) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	// Cobra default: false jika tidak di-set, jadi cek ENV jika flag tidak di-set
	if cmd.Flags().Changed(flagName) {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		env = strings.ToLower(env)
		return env == "1" || env == "true" || env == "yes"
	}
	return defaultVal
}
