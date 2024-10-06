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
	Long: `'substation fmt' formats configuration files.
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
		// Default to current directory if no path is provided.
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Catches an edge case where the user is looking for help.
		if path == "help" {
			fmt.Printf("warning: \"%s\" matched no files\n", path)
			return nil
		}

		write, err := cmd.Flags().GetBool("write")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		return formatPath(path, write, recursive)
	},
}

func formatPath(arg string, write, recursive bool) error {
	// Handle cases where the path is a file.
	ext := filepath.Ext(arg)
	if ext == ".jsonnet" || ext == ".libsonnet" {
		return formatFile(arg, write)
	}

	if err := filepath.WalkDir(arg, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if !recursive && path != arg {
				return filepath.SkipDir
			}

			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".jsonnet" && ext != ".libsonnet" {
			return nil
		}

		return formatFile(path, write)
	}); err != nil {
		return err
	}

	return nil
}

func formatFile(arg string, write bool) error {
	content, err := os.ReadFile(arg)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(arg, string(content), formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if !write {
		fmt.Println(formatted)

		return nil
	}

	if err := os.WriteFile(arg, []byte(formatted), 0o644); err != nil {
		return err
	}

	fmt.Println(arg)

	return nil
}
