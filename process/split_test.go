package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var splitByteTests = []struct {
	name     string
	proc     Split
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		Split{
			Options: SplitOptions{
				Separator: ".",
			},
			InputKey:  "split",
			OutputKey: "split",
		},
		[]byte(`{"split":"foo.bar"}`),
		[]byte(`{"split":["foo","bar"]}`),
		nil,
	},
	// the test case below is invalid because the Byter
	// cannot split a single item into multiple items
	{
		"invalid settings",
		Split{
			Options: SplitOptions{
				Separator: ".",
			},
		},
		[]byte(`foo.bar`),
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestSplitByte(t *testing.T) {
	ctx := context.TODO()
	for _, test := range splitByteTests {
		processed, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		expected := test.expected
		if c := bytes.Compare(expected, processed); c != 0 {
			t.Logf("expected %s, got %s", expected, string(processed))
			t.Fail()
		}
	}
}

func benchmarkSplitByte(b *testing.B, byter Split, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, data)
	}
}

func BenchmarkSplitByte(b *testing.B) {
	for _, test := range splitByteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkSplitByte(b, test.proc, test.test)
			},
		)
	}
}

var splitSliceTests = []struct {
	name     string
	proc     Split
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"data",
		Split{
			Options: SplitOptions{
				Separator: `\n`,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		nil,
	},
	{
		"invalid settings",
		Split{
			InputKey: "split",
			Options: SplitOptions{
				Separator: ".",
			},
		},
		[][]byte{},
		[][]byte{},
		ProcessorInvalidSettings,
	},
}

func TestSplitSlice(t *testing.T) {
	ctx := context.TODO()
	for _, test := range splitSliceTests {
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
				t.Logf("expected %s, got %s", expected, string(processed))
				t.Fail()
			}
		}
	}
}

func benchmarkSplitSlice(b *testing.B, slicer Split, slice [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, slice)
	}
}

func BenchmarkSplitSlice(b *testing.B) {
	for _, test := range splitSliceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkSplitSlice(b, test.proc, test.test)
			},
		)
	}
}
