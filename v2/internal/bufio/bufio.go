package bufio

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"strconv"

	"github.com/brexhq/substation/v2/internal/media"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
)

// MediaTypes natively supported by the bufio
// scanner.
var MediaTypes = []string{
	"application/x-bzip2",
	"application/x-gzip",
	"application/x-zstd",
	"application/x-snappy-framed",
	"text/plain; charset=utf-8",
}

// NewScanner returns a new
func NewScanner() *scanner {
	return &scanner{}
}

// scanner wraps bufio.scanner and provides methods for scanning data. The caller is responsible
// for checking errors and closing the scanner when it is no longer needed.
type scanner struct {
	*bufio.Scanner
	// openHandles contains all handles that must be closed after scanning is complete.
	openHandles []io.ReadCloser
}

// ReadFile inspects, decompresses, and reads an open file into the scanner. These file compression
// formats are optionally supported:
//   - bzip2 (https://en.wikipedia.org/wiki/Bzip2)
//   - gzip (https://en.wikipedia.org/wiki/Gzip)
//   - snappy (https://en.wikipedia.org/wiki/Snappy_(compression))
//   - zstd (https://en.wikipedia.org/wiki/Zstandard)
func (s *scanner) ReadFile(file *os.File) error {
	var reader io.ReadCloser
	s.openHandles = append(s.openHandles, file)

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	mediaType, err := media.File(file)
	if err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	switch mediaType {
	case "application/x-bzip2":
		reader = io.NopCloser(bzip2.NewReader(file))
		s.openHandles = append(s.openHandles, reader)
	case "application/x-gzip":
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}

		reader = gzipReader
		s.openHandles = append(s.openHandles, reader)
	case "application/zstd":
		zstdReader, err := zstd.NewReader(file)
		if err != nil {
			return err
		}

		reader = io.NopCloser(zstdReader)
		s.openHandles = append(s.openHandles, reader)
	case "application/x-snappy-framed":
		snappyReader := snappy.NewReader(file)
		reader = io.NopCloser(snappyReader)
	default:
		// file was previously added to openHandles
		reader = file
	}

	s.Scanner = bufio.NewScanner(reader)
	s.Scanner.Split(bufio.ScanLines)

	// Each line has a default capacity of 64 KB and a variable maximum capacity
	// (defaults to 128 MB).
	b := make([]byte, bufio.MaxScanTokenSize)
	if mem, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE"); !ok {
		s.Scanner.Buffer(b, (1000 * 1000 * 128))
	} else {
		m, _ := strconv.ParseFloat(mem, 64)
		// For AWS Lambda, the max capacity is 80% of the function's memory.
		s.Scanner.Buffer(b, 1000000*int(m*0.8))
	}

	return nil
}

func (s *scanner) Err() error {
	if err := s.Scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Close closes all open handles.
func (s *scanner) Close() error {
	for _, h := range s.openHandles {
		if err := h.Close(); err != nil {
			return err
		}
	}

	return nil
}
