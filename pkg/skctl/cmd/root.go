package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skctl",
	Short: "skctl controls the Skafos service mesh control plane",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetPrefix("error: ")
	log.SetFlags(0)
}
