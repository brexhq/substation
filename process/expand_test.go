package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var expandTests = []struct {
	name     string
	proc     Expand
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON",
		Expand{
			InputKey: "expand",
		},
		[]byte(`{"expand":[{"foo":"bar"}]}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"JSON extra key",
		Expand{
			InputKey: "expand",
		},
		[]byte(`{"expand":[{"foo":"bar"},{"quux":"corge"}],"baz":"qux"}`),
		[][]byte{
			[]byte(`{"foo":"bar","baz":"qux"}`),
			[]byte(`{"quux":"corge","baz":"qux"}`),
		},
		nil,
	},
	{
		"invalid settings",
		Expand{},
		[]byte{},
		[][]byte{},
		ProcessorInvalidSettings,
	},
}

func TestExpand(t *testing.T) {
	ctx := context.TODO()
	for _, test := range expandTests {
		slice := make([]config.Capsule, 1, 1)
		cap := config.NewCapsule()
		cap.SetData(test.test)
		slice[0] = cap

		res, err := test.proc.ApplyBatch(ctx, slice)
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

func benchmarkExpandCapSlice(b *testing.B, slicer Expand, slice []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.ApplyBatch(ctx, slice)
	}
}

func BenchmarkExpandCapSlice(b *testing.B) {
	for _, test := range expandTests {
		slice := make([]config.Capsule, 1, 1)
		cap := config.NewCapsule()
		cap.SetData(test.test)
		slice[0] = cap

		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkExpandCapSlice(b, test.proc, slice)
			},
		)
	}
}
