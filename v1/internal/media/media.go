// package media provides capabilities for inspecting the content of data and identifying its media (Multipurpose Internet Mail Extensions, MIME) type.
package media

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

// Bytes returns the media type of a byte slice.
func Bytes(b []byte) string {
	switch {
	// http.DetectContentType cannot detect bzip2
	case bytes.HasPrefix(b, []byte("\x42\x5a\x68")):
		return "application/x-bzip2"
	// http.DetectContentType occasionally (rarely) generates false positive matches for application/vnd.ms-fontobject when the bytes are application/x-gzip
	case bytes.HasPrefix(b, []byte("\x1f\x8b\x08")):
		return "application/x-gzip"
	// http.DetectContentType cannot detect zstd
	case bytes.HasPrefix(b, []byte("\x28\xb5\x2f\xfd")):
		return "application/x-zstd"
	// http.DetectContentType cannot detect snappy
	case bytes.HasPrefix(b, []byte("\xff\x06\x00\x00\x73\x4e\x61\x50\x70\x59")):
		return "application/x-snappy-framed"
	default:
		return http.DetectContentType(b)
	}
}

// File returns the media type of an open file. The caller is responsible for resetting the position of the file.
func File(f *os.File) (string, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return "", fmt.Errorf("media file: %v", err)
	}

	// http.DetectContentType reads the first 512 bytes of data
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", fmt.Errorf("media file: %v", err)
	}

	// buf is truncated to avoid false positives due to missing bytes
	buf = buf[:n]

	return Bytes(buf), nil
}
