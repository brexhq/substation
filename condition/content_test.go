package condition

import (
	"testing"
)

func TestContent(t *testing.T) {
	var tests = []struct {
		inspector Content
		test      []byte
		expected  bool
	}{
		// matching Gzip against valid Gzip header
		{
			Content{
				Type: "application/x-gzip",
			},
			[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
			true,
		},
		// matching Gzip against invalid Gzip header (bytes swapped)
		{
			Content{
				Type: "application/x-gzip",
			},
			[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
			false,
		},
		// matching Gzip against invalid Gzip header (bytes swapped) with negation
		{
			Content{
				Type:   "application/x-gzip",
				Negate: true,
			},
			[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
			true,
		},
		// matching Zip against valid Zip header
		{
			Content{
				Type: "application/zip",
			},
			[]byte{80, 75, 03, 04},
			true,
		},
		// matching Gzip against valid Zip header
		{
			Content{
				Type: "application/zip",
			},
			[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
			false,
		},
		// matching Zip against invalid Zip header (bytes swapped)
		{
			Content{
				Type: "application/zip",
			},
			[]byte{04, 75, 03, 80},
			false,
		},
	}

	for _, testing := range tests {
		check, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v", testing.expected, check)
			t.Fail()
		}
	}
}
