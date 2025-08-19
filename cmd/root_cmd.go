package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sfDBTools",
	Short: "sfDBTools CLI",
}

func Execute() error {
	return rootCmd.Execute()
}
