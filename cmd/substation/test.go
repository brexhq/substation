package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
	"github.com/spf13/cobra"
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
}

func test(ctx context.Context, file string) error {
	start := time.Now()

	fi, err := os.Open(file)
	if err != nil {
		return err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(fi).Decode(&cfg); err != nil {
		return err
	}

	// These configurations are not valid.
	if len(cfg.Transforms) == 0 {
		return nil
	}

	if len(cfg.Tests) == 0 {
		fmt.Printf("?\t%s\t[no tests]\n", file)

		return nil
	}

	var failedFile bool // Tracks if any test in a file failed.
	for _, test := range cfg.Tests {
		// setup creates the test environment.
		setup, err := substation.New(ctx, substation.Config{
			Transforms: test.Transforms,
		})
		if err != nil {
			return err
		}

		sMsgs, err := setup.Transform(ctx, message.New().AsControl())
		if err != nil {
			fmt.Printf("?\t%s\t[tf err]\n", file)

			//nolint:nilerr  // config errors should not disrupt the test.
			return nil
		}

		cnd, err := condition.New(ctx, test.Condition)
		if err != nil {
			return err
		}

		for _, msg := range sMsgs {
			if msg.IsControl() {
				continue
			}

			// tester contains the config that will be tested.
			// This has to be done for every message to ensure
			// that there is no state shared between tests.
			tester, err := substation.New(ctx, cfg.Config)
			if err != nil {
				return err
			}

			tMsgs, err := tester.Transform(ctx, msg)
			if err != nil {
				fmt.Printf("?\t%s\t[tf err]\n", file)

				//nolint:nilerr  // config errors should not disrupt the test.
				return nil
			}

			for _, msg := range tMsgs {
				if msg.IsControl() {
					continue
				}

				ok, err := cnd.Condition(ctx, msg)
				if err != nil {
					fmt.Printf("?\t%s\t[cnd err]\n", file)

					//nolint:nilerr  // config errors should not disrupt the test.
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
	}

	if failedFile {
		fmt.Printf("FAIL\t%s\t%s\t\n", file, time.Since(start).Round(time.Microsecond))
	} else {
		fmt.Printf("ok\t%s\t%s\t\n", file, time.Since(start).Round(time.Microsecond))
	}

	return nil
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

The current directory is used if no arg is provided.

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
      condition: sub.cnd.all([
        sub.cnd.str.eq({ object: {source_key: 'x'}, value: 'true' }),
      ])
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
`,
	// Examples:
	//  substation test [-R]
	//  substation test [-R] /path/to/configs
	// 	substation test /path/to/config.json
	Example: `  substation test [-R]
  substation test [-R] /path/to/configs
  substation test /path/to/config.json
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background() // This doesn't need to be canceled.

		// 'test' defaults to the current directory.
		arg, err := os.Getwd()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			arg = args[0]
		}

		fi, err := os.Stat(arg)
		if err != nil {
			return err
		}

		// If the arg is a file, then test that file. If the
		// arg is also not a directory, then return a warning.
		if !fi.IsDir() && filepath.Ext(arg) == ".json" {
			if err := test(ctx, arg); err != nil {
				return err
			}

			return nil
		} else if !fi.IsDir() {
			fmt.Printf("warning: \"%s\" matched no files\n", arg)

			return nil
		}

		var entries []string
		// Walk to get all files in the directory that end with `.json`.
		// These are assumed to be Substation configuration files, and
		// are validated before attempting to run tests.
		if err := filepath.WalkDir(arg, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip directories, except the one provided as an argument, if
			// the 'recursive' flag is not set.
			if d.IsDir() && path != arg && !cmd.Flag("recursive").Changed {
				return filepath.SkipDir
			}

			if filepath.Ext(path) == ".json" {
				entries = append(entries, path)
			}

			return nil
		}); err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Printf("warning: \"%s\" matched no files\n", arg)

			return nil
		}

		for _, entry := range entries {
			if err := test(ctx, entry); err != nil {
				return err
			}
		}

		return nil
	},
}
