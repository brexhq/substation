package media_test

import (
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/media"
)

func ExampleBytes() {
	b := []byte("\x42\x5a\x68")
	mediaType := media.Bytes(b)

	fmt.Println(mediaType)
	// Output: application/x-bzip2
}

func ExampleFile() {
	// temp file is used to simulate an open file and must be removed after the test completes
	file, _ := os.CreateTemp("", "substation")
	defer os.Remove(file.Name())
	defer file.Close()

	_, _ = file.Write([]byte("\x42\x5a\x68"))

	// media.File moves the file offset to zero
	mediaType, err := media.File(file)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(mediaType)
	// Output: application/x-bzip2
}

func Example_switch() {
	bytes := [][]byte{
		// application/x-bzip2
		[]byte("\x42\x5a\x68"),
		// application/x-gzip
		[]byte("\x1f\x8b\x08"),
		// text/html
		[]byte("\x3c\x68\x74\x6d\x6c\x3e"),
	}

	for _, b := range bytes {
		// use a switch statement to contextually distribute data to other functions
		switch media.Bytes(b) {
		case "application/x-bzip2":
			continue
			// bzip2(b)
		case "application/x-gzip":
			continue
			// gzip(b)
		case "text/html; charset=utf-8":
			continue
			// html(b)
		default:
			continue
		}
	}
}
