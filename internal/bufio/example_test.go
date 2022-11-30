package bufio_test

import (
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/bufio"
)

func ExampleNewScanner_setup() {
	s := bufio.NewScanner()
	defer s.Close()

	// sets maximum capacity to 1024 bytes
	// defaults to 65536 bytes
	s.SetCapacity(1024)

	// sets method to "bytes"
	// defaults to "text"
	if err := s.SetMethod("bytes"); err != nil {
		// handle error
		panic(err)
	}
}

func ExampleNewScanner_readFile() {
	// temp file is used to simulate an open file and must be removed after the test completes
	file, _ := os.CreateTemp("", "substation")
	defer os.Remove(file.Name())

	_, _ = file.Write([]byte("foo\nbar\nbaz"))

	// scanner closes all open handles, including the open file
	s := bufio.NewScanner()
	defer s.Close()

	// scanner automatically decompresses file and chooses appropriate scan method (default is "text")
	if err := s.ReadFile(file); err != nil {
		// handle error
		panic(err)
	}

	for s.Scan() {
		switch s.Method() {
		case "bytes":
			fmt.Println(s.Bytes())
		case "text":
			fmt.Println(s.Text())
		}
	}

	// Output:
	// foo
	// bar
	// baz
}
