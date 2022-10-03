package process

import (
	"bytes"
	"context"
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
		"data",
		Expand{},
		[]byte(`[{"foo":"bar"},{"quux":"corge"}]`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"quux":"corge"}`),
		},
		nil,
	},
}

func TestExpand(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range expandTests {
		slice := make([]config.Capsule, 1)
		capsule.SetData(test.test)
		slice[0] = capsule

		result, err := test.proc.ApplyBatch(ctx, slice)
		if err != nil {
			t.Error(err)
		}

		for i, res := range result {
			expected := test.expected[i]
			if !bytes.Equal(expected, res.Data()) {
				t.Errorf("expected %s, got %s", expected, res)
			}
		}
	}
}

func benchmarkExpand(b *testing.B, slicer Expand, slice []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = slicer.ApplyBatch(ctx, slice)
	}
}

func BenchmarkExpand(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range expandTests {
		slice := make([]config.Capsule, 1)
		capsule.SetData(test.test)
		slice[0] = capsule

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkExpand(b, test.proc, slice)
			},
		)
	}
}
