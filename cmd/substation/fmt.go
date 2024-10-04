package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.PersistentFlags().BoolP("write", "w", false, "write result to (source) file instead of stdout")
	fmtCmd.PersistentFlags().BoolP("recursive", "R", false, "recursively format all files")
}

var fmtCmd = &cobra.Command{
	Use:   "fmt [path]",
	Short: "format configs",
	Long: `'substation fmt' formats Jsonnet files.
It prints the formatted output to stdout by default.
Use the --write flag to update the files in-place.

The command can format a single file or a directory.
Use the --recursive flag to format all files in a directory and its subdirectories.

Supported file extensions: .jsonnet, .libsonnet`,
	Example: `  substation fmt config.jsonnet
  substation fmt -w config.jsonnet
  substation fmt -R /path/to/configs
  substation fmt -w -R /path/to/configs`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to current directory if no path is provided
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		write, _ := cmd.Flags().GetBool("write")
		recursive, _ := cmd.Flags().GetBool("recursive")

		return formatPath(path, write, recursive)
	},
}

func formatPath(path string, write, recursive bool) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return formatFile(path, write)
	}

	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".jsonnet" && ext != ".libsonnet" {
			return nil
		}
		return formatFile(path, write)
	})
}

func formatFile(path string, write bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(path, string(content), formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if write {
		err = os.WriteFile(path, []byte(formatted), 0o644)
		if err != nil {
			return err
		}
		fmt.Println(path)
	} else {
		fmt.Println(formatted)
	}

	return nil
}
