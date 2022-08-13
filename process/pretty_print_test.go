package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
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
		var caps []config.Capsule
		cap := config.NewCapsule()
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		res, err := test.proc.ApplyBatch(ctx, caps)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		for i, processed := range res {
			expected := test.expected[i]
			if c := bytes.Compare(expected, processed.GetData()); c != 0 {
				t.Logf("expected %s, got %s", expected, processed)
				t.Fail()
			}
		}
	}
}

func benchmarkPrettyPrintSlice(b *testing.B, batcher PrettyPrint, caps []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		batcher.ApplyBatch(ctx, caps)
	}
}

func BenchmarkPrettyPrintSlice(b *testing.B) {
	for _, test := range prettyPrintSliceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				caps := make([]config.Capsule, 1, 1)
				for _, t := range test.test {
					cap := config.NewCapsule()
					cap.SetData(t)
					caps = append(caps, cap)
				}

				benchmarkPrettyPrintSlice(b, test.proc, caps)
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
		cap := config.NewCapsule()
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(test.expected, res.GetData()); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkPrettyPrintByte(b *testing.B, proc PrettyPrint, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		proc.Apply(ctx, cap)
	}
}

func BenchmarkPrettyPrintByte(b *testing.B) {
	for _, test := range prettyPrintByteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkPrettyPrintByte(b, test.proc, cap)
			},
		)
	}
}
