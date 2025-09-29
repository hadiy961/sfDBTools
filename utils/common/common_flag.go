package common

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

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
