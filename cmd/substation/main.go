package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "substation",
	Long: "'substation' is a tool for managing Substation configurations.",
}

func init() {
	// Hides the 'completion' command.
	rootCmd.AddCommand(&cobra.Command{
		Use:    "completion",
		Short:  "generate the autocompletion script for the specified shell",
		Hidden: true,
	})

	// Hides the 'help' command.
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

// compileFile returns JSON from a Jsonnet file.
func compileFile(f string, extVars map[string]string) (string, error) {
	vm := jsonnet.MakeVM()
	for k, v := range extVars {
		vm.ExtVar(k, v)
	}

	res, err := vm.EvaluateFile(f)
	if err != nil {
		return "", err
	}

	return res, nil
}

// compileStr returns JSON from a Jsonnet string.
func compileStr(m string, extVars map[string]string) (string, error) {
	vm := jsonnet.MakeVM()
	for k, v := range extVars {
		vm.ExtVar(k, v)
	}

	res, err := vm.EvaluateAnonymousSnippet("snippet", m)
	if err != nil {
		return "", err
	}

	return res, nil
}

// pathVars returns the directory and file name of a file path.
func pathVars(p string) (string, string) {
	dir, fn := filepath.Split(p)
	ext := filepath.Ext(fn)
	fn = strings.TrimSuffix(fn, ext)

	return dir, fn
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
