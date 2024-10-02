package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "substation",
	Long: "'substation' is a tool for managing Substation configurations.",
}

func init() {
	// Hides the 'completion' command.
	rootCmd.AddCommand(&cobra.Command{
		Use:    "completion",
		Short:  "generate the autocompletion script for the specified shell",
		Hidden: true,
	})

	// Hides the 'help' command.
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
