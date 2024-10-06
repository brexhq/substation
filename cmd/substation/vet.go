package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/brexhq/substation/v2"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(vetCmd)
	vetCmd.PersistentFlags().BoolP("recursive", "R", false, "recursively vet all files")
	vetCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
}

// vetTransformRe captures the transform ID from a Substation error message.
// Example: `transform 324f1035-10a51b9a: object_target_key: missing required option` -> `324f1035-10a51b9a`
var vetTransformRe = regexp.MustCompile(`transform ([a-f0-9-]+):`)

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
			fmt.Printf("warning: \"%s\" matched no files\n", path)
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

	switch filepath.Ext(arg) {
	case ".jsonnet", ".libsonnet":
		mem, err := buildFile(arg, extVars)
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
		fi, err := fiConfig(arg)
		if err != nil {
			return err
		}

		cfg = fi
	}

	ctx := context.Background() // This doesn't need to be canceled.
	if _, err := substation.New(ctx, cfg.Config); err != nil {
		r := vetTransformRe.FindStringSubmatch(err.Error())

		// Cannot determine which transform failed. This should almost
		// never happen, unless something has modified the configuration
		// after it was compiled by Jsonnet.
		if len(r) == 0 {
			// Substation uses the transform name as a static transform ID.
			//
			// Example: `vet.json: transform hash_sha256: object_target_key: missing required option``
			fmt.Printf("%s: %v\n", arg, err)

			return nil
		}

		tfID := r[1] // The transform ID (e.g., `324f1035-10a51b9a`).
		for idx, tf := range cfg.Config.Transforms {
			if tf.Settings["id"] == tfID {
				// Example: `vet.jsonnet:3 transform 324f1035-10a51b9a: object_target_key: missing required option``
				fmt.Printf("%s:%d %v\n", arg, idx+1, err) // The line number is 1-based.
				fmt.Printf("\n    %s\n\n", tf)

				return nil
			}
		}
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
