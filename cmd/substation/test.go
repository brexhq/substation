package main

import (
	"context"
	"encoding/json"
	"errors"
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

type testConfig struct {
	Name       string          `json:"name"`
	Transforms []config.Config `json:"transforms"`
	Condition  config.Config   `json:"condition"`
}

// customConfig wraps the Substation config with support for tests.
type customConfig struct {
	substation.Config
	Tests []testConfig `json:"tests"`
}

// errTestFailed is returned when any test fails or cannot run. The
// results are already printed, so the error itself is not printed.
var errTestFailed = errors.New("test failed")

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

The command exits with a non-zero status if any test fails or
cannot run ('[error]'). A failing file does not stop the run, so
every file is still tested and reported. An error while loading or
parsing a config file still stops the run.

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

		failed, err := testPath(path, extStr, recursive)
		if err != nil {
			return err
		}

		if failed {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return errTestFailed
		}

		return nil
	},
}

// testPath tests every file in the path and reports whether any
// test failed or could not run. All files are tested even if an
// earlier file fails.
func testPath(arg string, extVars map[string]string, recursive bool) (bool, error) {
	// Handle cases where the path is a file.
	ext := filepath.Ext(arg)
	if ext == ".jsonnet" || ext == ".libsonnet" || ext == ".json" {
		return testFile(arg, extVars)
	}

	var failed bool
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

		f, err := testFile(path, extVars)
		if f {
			failed = true
		}

		return err
	}); err != nil {
		return failed, err
	}

	return failed, nil
}

// testFile tests a file and reports whether any test failed or
// could not run.
func testFile(arg string, extVars map[string]string) (bool, error) {
	var cfg customConfig

	// Switching directories is required to support relative imports.
	// The current directory is saved and restored after each test.
	wd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(wd)
	}()

	fileName := filepath.Base(arg)
	_ = os.Chdir(filepath.Dir(arg))

	// The default case is a no-op -- the file is skipped because
	// it cannot be tested.
	switch filepath.Ext(fileName) {
	case ".jsonnet", ".libsonnet":
		// If the Jsonnet cannot compile, then the file is invalid.
		mem, err := compileFile(fileName, extVars)
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			fmt.Fprint(os.Stderr, transformErrStr(err, arg, cfg))

			return true, nil
		}

		c, err := memConfig(mem)
		if err != nil {
			return false, err
		}

		cfg = c
	case ".json":
		c, err := fiConfig(fileName)
		if err != nil {
			return false, err
		}

		cfg = c
	}

	start := time.Now()

	// These are not valid Substation configs and must be ignored.
	if len(cfg.Transforms) == 0 {
		return false, nil
	}

	if len(cfg.Tests) == 0 {
		fmt.Printf("?\t%s\t[no tests]\n", arg)

		return false, nil
	}

	ctx := context.Background() // This doesn't need to be canceled.

	var failedFile bool // Tracks if any test in a file failed.
	for _, test := range cfg.Tests {
		// cnd asserts that the test is successful.
		cnd, err := condition.New(ctx, test.Condition)
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:tests:%s:condition", arg, test.Name), cfg))

			return true, nil
		}

		// setup creates the test messages that are tested.
		setup, err := substation.New(ctx, substation.Config{
			Transforms: test.Transforms,
		})
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			if len(transformRe.FindStringSubmatch(err.Error())) == 0 {
				fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:tests:%s", arg, test.Name), cfg))
			} else {
				fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:tests:%s:transforms", arg, test.Name), cfg))
			}

			return true, nil
		}

		// tester contains the config that will be tested.
		// This has to be done for each test to ensure
		// that there is no state shared between tests.
		tester, err := substation.New(ctx, cfg.Config)
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			if len(transformRe.FindStringSubmatch(err.Error())) == 0 {
				fmt.Fprint(os.Stderr, transformErrStr(err, arg, cfg))
			} else {
				fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:transforms", arg), cfg))
			}

			return true, nil
		}

		sMsgs, err := setup.Transform(ctx, message.New().AsControl())
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:tests:%s:transforms", arg, test.Name), cfg))

			return true, nil
		}

		tMsgs, err := tester.Transform(ctx, sMsgs...)
		if err != nil {
			fmt.Printf("?\t%s\t[error]\n", arg)
			fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:transforms", arg), cfg))

			return true, nil
		}

		for _, msg := range tMsgs {
			// Skip control messages because they contain no data.
			if msg.IsControl() {
				continue
			}

			ok, err := cnd.Condition(ctx, msg)
			if err != nil {
				fmt.Printf("?\t%s\t[error]\n", arg)
				fmt.Fprint(os.Stderr, transformErrStr(err, fmt.Sprintf("%s:tests:%s:condition", arg, test.Name), cfg))

				return true, nil
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

	return failedFile, nil
}
