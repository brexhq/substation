package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brexhq/substation/v2"
)

func init() {
	rootCmd.AddCommand(vetCmd)
	vetCmd.PersistentFlags().BoolP("recursive", "R", false, "recursively vet all files")
	vetCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
	vetCmd.Flags().SortFlags = false
	vetCmd.PersistentFlags().SortFlags = false
}

var vetCmd = &cobra.Command{
	Use:   "vet [path]",
	Short: "report config errors",
	Long: `'substation vet' reports errors in configuration files.

The 'recursive' flag can be used to vet all files in a 
directory, and the current directory is used if no arg is 
provided.

If an error is found, then the output always includes the
file path and error message. If the location of the error
is known, then the output also includes the line number
where the error occurred.

'vet' checks for two types of errors:
  - Jsonnet syntax errors
  - Substation configuration errors

Jsonnet syntax errors look like this, and include the line
number and column range where the error occurred:
  vet.jsonnet:19:36-38 Unknown variable: st

    sub.tf.obj.insert({obj: { trg: st.format('%s.-1', 'bar') }, value: 'baz'}),

Substation config errors look like this, and include the
line number where the error occurred in the 'transforms'
array:
  vet.jsonnet:3 transform 324f1035-10a51b9a: object_target_key: missing required option

    {"type":"hash_sha256","settings":{"id":"324f1035-10a51b9a","object":{"source_key":"foo"}}}
`,
	// Examples:
	//  substation vet [-R]
	//  substation vet [-R] /path/to/configs
	//  substation vet /path/to/config.json
	//  substation vet /path/to/config.jsonnet
	//  substation vet /path/to/my.libsonnet
	Example: `  substation vet [-R]
  substation vet [-R] /path/to/configs
  substation vet /path/to/config.json
  substation vet /path/to/config.jsonnet
  substation vet /path/to/my.libsonnet
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to current directory if no path is provided.
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Catches an edge case where the user is looking for help.
		if path == "help" {
			fmt.Printf("warning: use -h instead.\n")
			return nil
		}

		extStr, err := cmd.PersistentFlags().GetStringToString("ext-str")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		return vetPath(path, extStr, recursive)
	},
}

func vetFile(arg string, extVars map[string]string) error {
	// This uses the custom config from the `test` command.
	var cfg customConfig

	// Switching directories is required to support relative imports.
	// The current directory is saved and restored after each test.
	wd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(wd)
	}()

	fileName := filepath.Base(arg)
	_ = os.Chdir(filepath.Dir(arg))

	switch filepath.Ext(fileName) {
	case ".jsonnet", ".libsonnet":
		mem, err := compileFile(fileName, extVars)
		if err != nil {
			// This is an error in the Jsonnet syntax.
			// The line number and column range are included.
			//
			// Example: `vet.jsonnet:19:36-38 Unknown variable: st`
			fmt.Printf("%v\n", err)

			return nil
		}

		cfg, err = memConfig(mem)
		if err != nil {
			return err
		}
	case ".json":
		fi, err := fiConfig(fileName)
		if err != nil {
			return err
		}

		cfg = fi
	}

	ctx := context.Background() // This doesn't need to be canceled.
	if _, err := substation.New(ctx, cfg.Config); err != nil {
		if len(transformRe.FindStringSubmatch(err.Error())) == 0 {
			fmt.Fprint(os.Stderr, transformErrStr(err, arg, cfg))
		} else {
			fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:transforms", arg), cfg))
		}

		return nil
	}

	// No errors were found.
	//
	// Example: `vet.jsonnet`
	fmt.Printf("%s\n", arg)
	return nil
}

func vetPath(arg string, extVars map[string]string, recursive bool) error {
	fi, err := os.Stat(arg)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return vetFile(arg, extVars)
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
		if ext != ".jsonnet" && ext != ".libsonnet" && ext != ".json" {
			return nil
		}

		return vetFile(path, extVars)
	}); err != nil {
		return err
	}

	return nil
}
