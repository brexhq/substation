package media

import (
	"os"
	"testing"
)

var mediaTests = []struct {
	name     string
	test     []byte
	expected string
}{
	{
		"bzip2",
		[]byte("\x42\x5a\x68"),
		"application/x-bzip2",
	},
	{
		"gzip",
		[]byte("\x1f\x8b\x08"),
		"application/x-gzip",
	},
	{
		"zstd",
		[]byte("\x28\xb5\x2f\xfd"),
		"application/x-zstd",
	},
	{
		"snappy",
		[]byte("\xff\x06\x00\x00\x73\x4e\x61\x50\x70\x59"),
		"application/x-snappy-framed",
	},
}

func TestBytes(t *testing.T) {
	for _, test := range mediaTests {
		mediaType := Bytes(test.test)

		if mediaType != test.expected {
			t.Errorf("expected %s, got %s", test.expected, mediaType)
		}
	}
}

func benchmarkBytes(b *testing.B, test []byte) {
	for i := 0; i < b.N; i++ {
		_ = Bytes(test)
	}
}

func BenchmarkBytes(b *testing.B) {
	for _, test := range mediaTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkBytes(b, test.test)
			},
		)
	}
}

func TestFile(t *testing.T) {
	for _, test := range mediaTests {
		temp, err := os.CreateTemp("", "substation")
		if err != nil {
			t.Errorf("got error %v", err)
		}
		defer os.Remove(temp.Name())
		defer temp.Close()

		if _, err := temp.Write(test.test); err != nil {
			t.Errorf("got error %v", err)
		}

		mediaType, err := File(temp)
		if err != nil {
			t.Errorf("got error %v", err)
		}

		if mediaType != test.expected {
			t.Errorf("expected %s, got %s", test.expected, mediaType)
		}
	}
}

func benchmarkFile(b *testing.B, test *os.File) {
	for i := 0; i < b.N; i++ {
		_, _ = File(test)
	}
}

func BenchmarkFile(b *testing.B) {
	temp, _ := os.CreateTemp("", "substation")
	defer os.Remove(temp.Name())
	defer temp.Close()

	for _, test := range mediaTests {
		_, _ = temp.Seek(0, 0)
		_, _ = temp.Write(test.test)

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFile(b, temp)
			},
		)
	}
}

func FuzzFile(f *testing.F) {
	// Seed the fuzzer with initial test cases
	f.Add([]byte("foo"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		temp, err := os.CreateTemp("", "substation")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(temp.Name())
		defer temp.Close()

		_, err = temp.Write(data)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}

		_, err = temp.Seek(0, 0)
		if err != nil {
			t.Fatalf("failed to seek to beginning of file: %v", err)
		}

		mediaType, err := File(temp)
		if err != nil {
			if err.Error() == "media file: EOF" && len(data) == 0 {
				return
			}
			t.Errorf("got error %v", err)
		}

		// Optionally, you can add more checks on mediaType here
		_ = mediaType
	})
}
