// package bufio wraps the standard library's bufio package.
package bufio

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/media"
)

// errInvalidMethod is returned when an invalid scanner method is provided.
const errInvalidMethod = errors.Error("invalid scanner method")

// NewScanner returns a new scanner with default settings provided by scanCapacity and scanMethod.
func NewScanner() *scanner {
	s := &scanner{}
	s.size = scanCapacity()
	s.method = scanMethod()

	return s
}

/*
scanner wraps bufio.scanner and provides methods for scanning data. The caller is responsible for closing the scanner when it is no longer needed.
*/
type scanner struct {
	*bufio.Scanner
	// openHandles contains all handles that must be closed after scanning is complete.
	openHandles []io.ReadCloser
	// size defines the maximum size used to buffer a token during scanning.
	size int
	// method defines how tokens should be generated when calling the scanner.
	method string
}

// Capacity returns the capacity setting for the scanner.
func (s *scanner) Capacity() int {
	return s.size
}

// SetCapacity sets the maximum buffer size that may be allocated during scanning. This method overrides the default capacity that is set when the scanner is created.
func (s *scanner) SetCapacity(capacity int) {
	s.size = capacity
}

// Method returns the method setting for the scanner.
func (s *scanner) Method() string {
	return s.method
}

// SetMethod sets the method for the scanner. This method overrides the default method that is set when the scanner is created.
func (s *scanner) SetMethod(m string) error {
	if m == "bytes" || m == "text" {
		s.method = m

		return nil
	}

	return fmt.Errorf("setmethod: %v", errInvalidMethod)
}

/*
ReadFile inspects, contextually decompresses, and reads an open file into the scanner. These file formats are optionally supported:

- bzip2 (https://en.wikipedia.org/wiki/Bzip2)

- gzip (https://en.wikipedia.org/wiki/Gzip)
*/
func (s *scanner) ReadFile(file *os.File) error {
	var reader io.ReadCloser
	s.openHandles = append(s.openHandles, file)

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("readfile: %v", err)
	}

	mediaType, err := media.File(file)
	if err != nil {
		return fmt.Errorf("readfile: %v", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("readfile: %v", err)
	}

	switch mediaType {
	case "application/x-bzip2":
		reader = io.NopCloser(bzip2.NewReader(file))
		s.openHandles = append(s.openHandles, reader)
	case "application/x-gzip":
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("readopenfile: %v", err)
		}

		reader = gzipReader
		s.openHandles = append(s.openHandles, reader)
	default:
		// file was previously added to openHandles
		reader = file
	}

	s.Scanner = bufio.NewScanner(reader)
	s.Scanner.Buffer([]byte{}, s.size)

	return nil
}

// Close closes all open handles.
func (s *scanner) Close() error {
	for _, h := range s.openHandles {
		if err := h.Close(); err != nil {
			return fmt.Errorf("close: %v", err)
		}
	}

	return nil
}

/*
scanCapacity retrieves a value from the SUBSTATION_SCAN_CAPACITY environment variable. This impacts the maximum size of each token (e.g., line in a text file) that is buffered (read) by bufio scanners that are used throughout the application to read files. More information about capacity can be found in https://pkg.go.dev/bufio#pkg-constants.

If the environment variable is missing, then the default capacity is the value from the bufio package (approximately 65.5 KB).
*/
func scanCapacity() int {
	maxSize := bufio.MaxScanTokenSize

	if val, found := os.LookupEnv("SUBSTATION_SCAN_CAPACITY"); found {
		v, err := strconv.Atoi(val)
		if err != nil {
			return maxSize
		}

		return v
	}

	return maxSize
}

/*
scanMethod retrieves a value from the SUBSTATION_SCAN_METHOD environment variable. This impacts the read behavior of bufio scanners that are used throughout the application to read files. The options for this variable are:

- "bytes" (https://pkg.go.dev/bufio#Scanner.Bytes)

- "text" (https://pkg.go.dev/bufio#Scanner.Text)

If the environment variable is missing, then the default method is "text".
*/
func scanMethod() string {
	if val, found := os.LookupEnv("SUBSTATION_SCAN_METHOD"); found {
		if val == "bytes" || val == "text" {
			return val
		}
	}

	return "text"
}
