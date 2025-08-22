package main

import (
	"os"

	"sfDBTools_new/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
