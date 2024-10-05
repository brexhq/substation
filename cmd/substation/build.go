package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().BoolP("recursive", "R", false, "recursively build all files")
	buildCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
}

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "build configs",
	Long: `'substation build' compiles configuration files.

The 'recursive' flag can be used to build all files in a directory,
and the current directory is used if no arg is provided.`,
	Example: `  substation build [-R]
  substation build [-R] /path/to/configs
  substation build config.jsonnet
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to current directory if no path is provided
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		m, err := cmd.PersistentFlags().GetStringToString("ext-str")
		if err != nil {
			return err
		}

		r, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		return buildPath(path, m, r)
	},
}

func buildPath(arg string, extVars map[string]string, recursive bool) error {
	// Handle cases where the path is a file.
	//
	// Only `.jsonnet` files are built.
	if filepath.Ext(arg) == ".jsonnet" {
		mem, err := buildFile(arg, extVars)
		if err != nil {
			return err
		}

		dir, fname := pathVars(arg)
		if err := os.WriteFile(filepath.Join(dir, fname)+".json", []byte(mem), 0o644); err != nil {
			return err
		}

		// Print the file that was built.
		fmt.Println(arg)

		return nil
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

		// Only `.jsonnet` files are built.
		if filepath.Ext(path) != ".jsonnet" {
			return nil
		}

		mem, err := buildFile(path, extVars)
		if err != nil {
			return err
		}

		dir, fname := pathVars(path)
		if err := os.WriteFile(filepath.Join(dir, fname)+".json", []byte(mem), 0o644); err != nil {
			return err
		}

		// Print the file that was built.
		fmt.Println(path)

		return nil
	}); err != nil {
		return err
	}

	return nil
}
