package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var prettyPrintSliceTests = []struct {
	name     string
	proc     PrettyPrint
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"from",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "from",
			},
		},
		[][]byte{
			[]byte(`{
				"foo":"bar"
				}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"from",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "from",
			},
		},
		[][]byte{
			[]byte(`{`),
			[]byte(`"foo":"bar",`),
			[]byte(`"baz": {`),
			[]byte(`	"qux": "corge"`),
			[]byte(`}`),
			[]byte(`}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar","baz":{"qux":"corge"}}`),
		},
		nil,
	},
	{
		"to",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "to",
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		[][]byte{
			[]byte(`{
  "foo": "bar"
}
`),
		},
		nil,
	},
	{
		"invalid settings",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "foo",
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		[][]byte{},
		ProcessorInvalidSettings,
	},
	{
		"unbalanced brackets",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "from",
			},
		},
		[][]byte{
			[]byte(`{{`),
			[]byte(`"foo":"bar"`),
			[]byte(`}`),
		},
		[][]byte{},
		PrettyPrintUnbalancedBrackets,
	},
}

func TestPrettyPrintSlice(t *testing.T) {
	ctx := context.TODO()
	for _, test := range prettyPrintSliceTests {
		res, err := test.proc.Slice(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		for i, processed := range res {
			expected := test.expected[i]
			if c := bytes.Compare(expected, processed); c != 0 {
				t.Logf("expected %s, got %s", expected, processed)
				t.Fail()
			}
		}
	}
}

func benchmarkPrettyPrintSlice(b *testing.B, slicer PrettyPrint, slice [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, slice)
	}
}

func BenchmarkPrettyPrintSlice(b *testing.B) {
	for _, test := range prettyPrintSliceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkPrettyPrintSlice(b, test.proc, test.test)
			},
		)
	}
}

var prettyPrintByteTests = []struct {
	name     string
	proc     PrettyPrint
	test     []byte
	expected []byte
	err      error
}{
	{
		"to",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "to",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{
  "foo": "bar"
}
`),
		nil,
	},
	// PrettyPrint from is not supported in Byter
	{
		"invalid settings",
		PrettyPrint{
			Options: PrettyPrintOptions{
				Direction: "from",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestPrettyPrintByte(t *testing.T) {
	ctx := context.TODO()
	for _, test := range prettyPrintByteTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(test.expected, res); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkPrettyPrintByte(b *testing.B, byter PrettyPrint, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, data)
	}
}

func BenchmarkPrettyPrintByte(b *testing.B) {
	for _, test := range prettyPrintByteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkPrettyPrintByte(b, test.proc, test.test)
			},
		)
	}
}
