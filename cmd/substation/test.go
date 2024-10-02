package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/condition"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/file"
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

// getConfig contextually retrieves a Substation configuration.
func getConfig(ctx context.Context, cfg string) (io.Reader, error) {
	path, err := file.Get(ctx, cfg)
	defer os.Remove(path)

	if err != nil {
		return nil, err
	}

	conf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer conf.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, conf); err != nil {
		return nil, err
	}

	return buf, nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test [path to configs]",
	Short: "test configs",
	Long: `'substation test' runs all tests in configuration files.
It prints a summary of the test results in the format:

  ok	path/to/config1.json 	2ms
  ?	path/to/config2.json 	[no tests]
  FAIL 	path/to/config3.json 	1ms
  ...

The current directory is used if no path is provided.

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
        sub.tf.utility.message({ value: {a: true} }),
      ],
      // Checks if key 'x' == 'true'.
      condition: sub.cnd.all([
        sub.cnd.str.eq({ object: {source_key: 'x'}, value: 'true' }),
      ])
    },
    {
      name: 'my-failing-test',
      transforms: [
        sub.tf.utility.message({ value: {a: true} }),
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
	//  substation test
	//  substation test /path/to/configs/
	Example: `  substation test
  substation test /path/to/configs/
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 'test' defaults to the current directory.
		arg, err := os.Getwd()
		if err != nil {
			return err
		}

		// Validate if the arg is a directory.
		if len(args) > 0 {
			arg = args[0]

			fi, err := os.Stat(arg)
			if err != nil {
				return err
			}

			if !fi.IsDir() {
				// Error: 'test.json' is not a directory
				return fmt.Errorf("'%s' is not a directory\n", arg)
			}
		}

		var entries []string
		// Walk to get all files in the directory that end with `.json`.
		// These are assumed to be Substation configuration files, and
		// are validated before attempting to run tests.
		if err := filepath.WalkDir(arg, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
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

		ctx := context.Background() // This doesn't need to be canceled.
		for _, entry := range entries {
			start := time.Now()

			c, err := getConfig(ctx, entry)
			if err != nil {
				fmt.Printf("?\t%s\t[empty file]\n", entry)

				continue
			}

			cfg := customConfig{}
			if err := json.NewDecoder(c).Decode(&cfg); err != nil {
				return err
			}

			// These configurations are not valid.
			if len(cfg.Transforms) == 0 {
				continue
			}

			if len(cfg.Tests) == 0 {
				fmt.Printf("?\t%s\t[no tests]\n", entry)

				continue
			}

			// subT contains the config that will be tested.
			subT, err := substation.New(ctx, cfg.Config)
			if err != nil {
				return err
			}

			var failed bool
			for _, test := range cfg.Tests {
				// sub sets up the test environment.
				sub, err := substation.New(ctx, substation.Config{
					Transforms: test.Transforms,
				})
				if err != nil {
					return err
				}

				cnd, err := condition.New(ctx, test.Condition)
				if err != nil {
					return err
				}

				// ctrl message is used to trigger the test.
				msgs, err := sub.Transform(ctx, message.New().AsControl())
				if err != nil {
					return err
				}

				for _, msg := range msgs {
					subTMsgs, err := subT.Transform(ctx, msg)
					if err != nil {
						return err
					}

					// Every message produced by the config must be checked.
					for _, subTMsg := range subTMsgs {
						if subTMsg.IsControl() {
							continue
						}

						ok, err := cnd.Condition(ctx, subTMsg)
						if err != nil {
							return err
						}

						if !ok {
							fmt.Printf("%s\n%s\n%s\n",
								fmt.Sprintf("--- FAIL: %s", test.Name),
								fmt.Sprintf("    message:\t%s", subTMsg),
								fmt.Sprintf("    condition:\t%s", cnd),
							)

							failed = true
						}
					}
				}
			}

			if !failed {
				fmt.Printf("ok\t%s\t%s\t\n", entry, time.Since(start).Round(time.Millisecond))
			} else {
				fmt.Printf("FAIL\t%s\t%s\t\n", entry, time.Since(start).Round(time.Millisecond))
			}
		}

		return nil
	},
}
