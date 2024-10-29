package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

// customConfig wraps the Substation config with support for tests.
type customConfig struct {
	substation.Config

	Tests []struct {
		Name       string          `json:"name"`
		Transforms []config.Config `json:"transforms"`
		Condition  config.Config   `json:"condition"`
	} `json:"tests"`
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().BoolP("recursive", "R", false, "recursively test all files")
	testCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
	testCmd.Flags().SortFlags = false
	testCmd.PersistentFlags().SortFlags = false
}

func fiConfig(f string) (customConfig, error) {
	fi, err := os.Open(f)
	if err != nil {
		if err == io.EOF {
			return customConfig{}, nil
		}

		return customConfig{}, err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(fi).Decode(&cfg); err != nil {
		return customConfig{}, err
	}

	return cfg, nil
}

func memConfig(m string) (customConfig, error) {
	cfg := customConfig{}

	if err := json.Unmarshal([]byte(m), &cfg); err != nil {
		return customConfig{}, err
	}

	return cfg, nil
}

var testCmd = &cobra.Command{
	Use:   "test [path to configs]",
	Short: "test configs",
	Long: `'substation test' runs all tests in a configuration file.
It prints a summary of the test results in the format:

  ok	path/to/config1.json 	220µs
  ?	path/to/config2.json 	[no tests]
  FAIL 	path/to/config3.json 	349µs
  ...

If the file is not already compiled, then it is compiled before
testing ('.jsonnet', '.libsonnet' files are compiled to JSON).
The 'recursive' flag can be used to test all files in a directory,
and the current directory is used if no arg is provided.

Tests are executed individually against configured transforms. 
Each test executes on user-defined messages and is considered
successful if a condition returns true for every message.

For example, this config contains two tests:

{
  tests: [
    {
      name: 'my-passing-test',
      // Generates the test message '{"a": true}' which
      // is run through the configured transforms and
      // then checked against the condition.
      transforms: [
        sub.tf.test.message({ value: {a: true} }),
      ],
      // Checks if key 'x' == 'true'.
      condition: sub.cnd.str.eq({ object: {source_key: 'x'}, value: 'true' }),
    },
    {
      name: 'my-failing-test',
      transforms: [
        sub.tf.test.message({ value: {a: true} }),
      ],
      // Checks if key 'y' == 'true'.
      condition: sub.cnd.str.eq({ object: {source_key: 'y'}, value: 'true' }),
    },
  ],
  // Copies the value of key 'a' to key 'x'.
  transforms: [
    sub.tf.obj.cp({ object: { source_key: 'a', target_key: 'x' } }),
  ],
}

WARNING: It is not recommended to test any configs that mutate
production resources, such as any enrichment or send transforms.
`,
	// Examples:
	//  substation test [-R]
	//  substation test [-R] /path/to/configs
	//  substation test /path/to/config.json
	//  substation test /path/to/config.jsonnet
	//  substation test /path/to/my.libsonnet
	Example: `  substation test [-R]
  substation test [-R] /path/to/configs
  substation test /path/to/config.json
  substation test /path/to/config.jsonnet
  substation test /path/to/my.libsonnet
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		return testPath(path, extStr, recursive)
	},
}

func testPath(arg string, extVars map[string]string, recursive bool) error {
	// Handle cases where the path is a file.
	ext := filepath.Ext(arg)
	if ext == ".jsonnet" || ext == ".libsonnet" || ext == ".json" {
		return testFile(arg, extVars)
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

		return testFile(path, extVars)
	}); err != nil {
		return err
	}

	return nil
}

func testFile(arg string, extVars map[string]string) error {
	var cfg customConfig

	switch filepath.Ext(arg) {
	case ".jsonnet", ".libsonnet":
		// If the Jsonnet cannot compile, then the file is invalid.
		mem, err := compileFile(arg, extVars)
		if err != nil {
			fmt.Printf("?\t%s\t[config error]\n", arg)

			//nolint:nilerr  // errors should not disrupt the test.
			return nil
		}

		c, err := memConfig(mem)
		if err != nil {
			return err
		}

		cfg = c
	case ".json":
		c, err := fiConfig(arg)
		if err != nil {
			return err
		}

		cfg = c
	}

	start := time.Now()

	// These configurations are not valid.
	if len(cfg.Transforms) == 0 {
		return nil
	}

	if len(cfg.Tests) == 0 {
		fmt.Printf("?\t%s\t[no tests]\n", arg)

		return nil
	}

	ctx := context.Background() // This doesn't need to be canceled.

	var failedFile bool // Tracks if any test in a file failed.
	for _, test := range cfg.Tests {
		// cnd asserts that the test is successful.
		cnd, err := condition.New(ctx, test.Condition)
		if err != nil {
			fmt.Printf("FAIL\t%s\t[test error]\n", arg)

			//nolint:nilerr  // errors should not disrupt the test.
			return nil
		}

		// setup creates the test messages that are tested.
		setup, err := substation.New(ctx, substation.Config{
			Transforms: test.Transforms,
		})
		if err != nil {
			fmt.Printf("?\t%s\t[test error]\n", arg)

			//nolint:nilerr  // errors should not disrupt the test.
			return nil
		}

		// tester contains the config that will be tested.
		// This has to be done for each test to ensure
		// that there is no state shared between tests.
		tester, err := substation.New(ctx, cfg.Config)
		if err != nil {
			fmt.Printf("?\t%s\t[config error]\n", arg)

			//nolint:nilerr  // errors should not disrupt the test.
			return nil
		}

		sMsgs, err := setup.Transform(ctx, message.New().AsControl())
		if err != nil {
			fmt.Printf("?\t%s\t[test.transform error]\n", arg)
			fmt.Fprintf(os.Stderr, "\t\t%v\n", err)

			return nil
		}

		tMsgs, err := tester.Transform(ctx, sMsgs...)
		if err != nil {
			fmt.Printf("?\t%s\t[transform error]\n", arg)
			fmt.Fprintf(os.Stderr, "\t\t%v\n", err)

			return nil
		}

		for _, msg := range tMsgs {
			// Skip control messages because they contain no data.
			if msg.HasFlag(message.IsControl) {
				continue
			}

			ok, err := cnd.Condition(ctx, msg)
			if err != nil {
				fmt.Printf("?\t%s\t[test error]\n", arg)

				//nolint:nilerr  // errors should not disrupt the test.
				return nil
			}

			if !ok {
				fmt.Printf("%s\n%s\n%s\n",
					fmt.Sprintf("--- FAIL: %s", test.Name),
					fmt.Sprintf("    message:\t%s", msg),
					fmt.Sprintf("    condition:\t%s", cnd),
				)

				failedFile = true
			}
		}
	}

	if failedFile {
		fmt.Printf("FAIL\t%s\t%s\t\n", arg, time.Since(start).Round(time.Microsecond))
	} else {
		fmt.Printf("ok\t%s\t%s\t\n", arg, time.Since(start).Round(time.Microsecond))
	}

	return nil
}
