package base64

import (
	"bytes"
	"testing"
)

var decodeTests = []struct {
	name     string
	test     []byte
	expected []byte
}{
	{
		name:     "foo",
		test:     []byte(`Zm9v`),
		expected: []byte(`foo`),
	},
	{
		name:     "zlib",
		test:     []byte(`eJwFwCENAAAAgLC22Pd3LAYCggFF`),
		expected: []byte{120, 156, 5, 192, 33, 13, 0, 0, 0, 128, 176, 182, 216, 247, 119, 44, 6, 2, 130, 1, 69},
	},
}

func TestBase64Decode(t *testing.T) {
	for _, test := range decodeTests {
		result, err := Decode(test.test)
		if err != nil {
			t.Errorf("got error %v", err)
			return
		}

		if c := bytes.Compare(result, test.expected); c != 0 {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkBase64Decode(b *testing.B, test []byte) {
	for i := 0; i < b.N; i++ {
		_, _ = Decode(test)
	}
}

func BenchmarkBase64Decode(b *testing.B) {
	for _, test := range encodeTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkBase64Decode(b, test.test)
			},
		)
	}
}

var encodeTests = []struct {
	name     string
	test     []byte
	expected []byte
}{
	{
		name:     "foo",
		test:     []byte(`foo`),
		expected: []byte(`Zm9v`),
	},
	{
		name:     "zlib",
		test:     []byte{120, 156, 5, 192, 33, 13, 0, 0, 0, 128, 176, 182, 216, 247, 119, 44, 6, 2, 130, 1, 69},
		expected: []byte(`eJwFwCENAAAAgLC22Pd3LAYCggFF`),
	},
}

func TestBase64Encode(t *testing.T) {
	for _, test := range encodeTests {
		result := Encode(test.test)

		if c := bytes.Compare(result, test.expected); c != 0 {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkBase64Encode(b *testing.B, test []byte) {
	for i := 0; i < b.N; i++ {
		Encode(test)
	}
}

func BenchmarkBase64Encode(b *testing.B) {
	for _, test := range encodeTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkBase64Encode(b, test.test)
			},
		)
	}
}

func FuzzBase64Encode(f *testing.F) {
	// Seed the fuzzer with initial test cases
	for _, test := range encodeTests {
		f.Add(test.test)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Encode the input data
		result := Encode(data)

		// Decode the result to verify it matches the original input
		decoded, err := Decode(result)
		if err != nil {
			t.Errorf("failed to decode: %v", err)
		}

		if !bytes.Equal(data, decoded) {
			t.Errorf("expected %s, got %s", data, decoded)
		}
	})
}
