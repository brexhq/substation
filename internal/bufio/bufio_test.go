package bufio

import (
	"os"
	"testing"
)

func benchmarkScannerReadFile(b *testing.B, s *scanner, file *os.File) {
	for i := 0; i < b.N; i++ {
		_ = s.ReadFile(file)
		for s.Scan() {
			s.Text()
		}
	}
}

func BenchmarkScannerReadFile(b *testing.B) {
	file, _ := os.CreateTemp("", "substation")
	defer os.Remove(file.Name())

	_, _ = file.Write([]byte("foo\nbar\nbaz"))

	s := NewScanner()
	defer s.Close()

	b.Run("scanner_read_file",
		func(b *testing.B) {
			benchmarkScannerReadFile(b, s, file)
		},
	)
}

func FuzzScannerReadFile(f *testing.F) {
	// Seed the fuzzer with initial test cases
	f.Add([]byte("foo\nbar\nbaz"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		file, err := os.CreateTemp("", "substation")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(file.Name())

		_, err = file.Write(data)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}

		// Ensure the file is properly closed and flushed
		err = file.Close()
		if err != nil {
			t.Fatalf("failed to close temp file: %v", err)
		}

		// Reopen the file for reading
		file, err = os.Open(file.Name())
		if err != nil {
			t.Fatalf("failed to reopen temp file: %v", err)
		}
		defer file.Close()

		s := NewScanner()
		defer s.Close()

		err = s.ReadFile(file)
		if err != nil {
			if err.Error() == "media file: EOF" && len(data) == 0 {
				return
			}
			t.Fatalf("failed to read file: %v", err)
		}

		for s.Scan() {
			_ = s.Text()
		}

		if err := s.Err(); err != nil {
			t.Errorf("scanner error: %v", err)
		}
	})
}
