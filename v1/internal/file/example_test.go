package file_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/brexhq/substation/internal/file"
)

func ExampleGet_local() {
	// temp file is used to simulate an open file and must be removed after the test completes
	temp, _ := os.CreateTemp("", "substation")
	defer os.Remove(temp.Name())
	defer temp.Close()

	_, _ = temp.Write([]byte("foo\nbar\nbaz"))

	// a local copy of the file is created and must be removed when it's no longer needed, regardless of errors
	path, err := file.Get(context.TODO(), temp.Name())
	defer os.Remove(path)

	if err != nil {
		// handle err
		panic(err)
	}

	f, err := os.Open(path)
	if err != nil {
		// handle err
		panic(err)
	}

	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(string(buf))

	// Output:
	// foo
	// bar
	// baz
}

func ExampleGet_http() {
	location := "https://example.com"

	// a local copy of the HTTP body is created and must be removed when it's no longer needed, regardless of errors
	path, err := file.Get(context.TODO(), location)
	defer os.Remove(path)

	if err != nil {
		// handle err
		panic(err)
	}

	f, err := os.Open(path)
	if err != nil {
		// handle err
		panic(err)
	}

	defer f.Close()

	buf := make([]byte, 16)
	if _, err = f.Read(buf); err != nil {
		// handle err
		panic(err)
	}

	prefix := strings.HasPrefix(strings.ToUpper(string(buf)), "<!DOCTYPE")
	fmt.Println(prefix)

	// Output: true
}
