package condition

import (
	"testing"
)

var contentTests = []struct {
	name      string
	inspector Content
	test      []byte
	expected  bool
}{
	// matching Gzip against valid Gzip header
	{
		"gzip",
		Content{
			Type: "application/x-gzip",
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		true,
	},
	// matching Gzip against invalid Gzip header (bytes swapped)
	{
		"!gzip",
		Content{
			Type: "application/x-gzip",
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		false,
	},
	// matching Gzip against invalid Gzip header (bytes swapped) with negation
	{
		"gzip",
		Content{
			Type:   "application/x-gzip",
			Negate: true,
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		true,
	},
	// matching Zip against valid Zip header
	{
		"zip",
		Content{
			Type: "application/zip",
		},
		[]byte{80, 75, 03, 04},
		true,
	},
	// matching Gzip against valid Zip header
	{
		"!zip",
		Content{
			Type: "application/zip",
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		false,
	},
	// matching Zip against invalid Zip header (bytes swapped)
	{
		"!zip",
		Content{
			Type: "application/zip",
		},
		[]byte{04, 75, 03, 80},
		false,
	},
}

func TestContent(t *testing.T) {
	for _, testing := range contentTests {
		check, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v", testing.expected, check)
			t.Fail()
		}
	}
}

func benchmarkContentByte(b *testing.B, inspector Content, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkContentByte(b *testing.B) {
	for _, test := range contentTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkContentByte(b, test.inspector, test.test)
			},
		)
	}
}
