package bufio

import (
	"os"
	"testing"
)

func benchmarkScannerReadFile(b *testing.B, s *scanner, file *os.File) {
	for i := 0; i < b.N; i++ {
		s.ReadFile(file)
		for s.Scan() {
			s.Text()
		}
	}
}

func BenchmarkScannerReadFile(b *testing.B) {
	file, _ := os.CreateTemp("", "substation")
	defer os.Remove(file.Name())

	file.Write([]byte("foo\nbar\nbaz"))

	s := NewScanner()
	defer s.Close()

	b.Run("scanner_read_file",
		func(b *testing.B) {
			benchmarkScannerReadFile(b, s, file)
		},
	)
}
