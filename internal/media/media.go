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
	// bzip2 is undetected by http.DetectContentType
	case bytes.HasPrefix(b, []byte("\x42\x5a\x68")):
		return "application/x-bzip2"
	default:
		return http.DetectContentType(b)
	}
}

// File returns the media type of an open file. The caller is responsible for resetting the position of the file.
func File(f *os.File) (string, error) {
	f.Seek(0, 0)

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
