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
	buildCmd.Flags().SortFlags = false
	buildCmd.PersistentFlags().SortFlags = false
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

		// Catches an edge case where the user is looking for help.
		if path == "help" {
			fmt.Printf("warning: use -h instead.\n")
			return nil
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
		return buildFile(arg, extVars)
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

		return buildFile(path, extVars)
	}); err != nil {
		return err
	}

	return nil
}

func buildFile(arg string, extVars map[string]string) error {
	mem, err := compileFile(arg, extVars)
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
