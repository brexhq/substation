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
