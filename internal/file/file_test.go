package file_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/brexhq/substation/v2/internal/file"
)

func FuzzExampleGet_http(f *testing.F) {
	// Seed the fuzzer with initial test cases
	f.Add("https://example.com")
	// f.Add("https://invalid-url") // TODO (akline@brex.com): should add URL validation in file.Get

	f.Fuzz(func(t *testing.T, location string) {
		// a local copy of the HTTP body is created and must be removed when it's no longer needed, regardless of errors
		path, err := file.Get(context.TODO(), location)
		if err != nil {
			// handle err
			return
		}
		defer os.Remove(path)

		f, err := os.Open(path)
		if err != nil {
			// handle err
			return
		}
		defer f.Close()

		buf := make([]byte, 16)
		if _, err = f.Read(buf); err != nil {
			// handle err
			return
		}

		prefix := strings.HasPrefix(strings.ToUpper(string(buf)), "<!DOCTYPE")
		fmt.Println(prefix)
	})
}
