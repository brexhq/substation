package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var splitTests = []struct {
	name     string
	proc     split
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		split{
			process: process{
				Key:    "split",
				SetKey: "split",
			},
			Options: splitOptions{
				Separator: ".",
			},
		},
		[]byte(`{"split":"foo.bar"}`),
		[]byte(`{"split":["foo","bar"]}`),
		nil,
	},
}

func TestSplit(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range splitTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkSplit(b *testing.B, proc split, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Apply(ctx, capsule)
	}
}

func BenchmarkSplit(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range splitTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkSplit(b, test.proc, capsule)
			},
		)
	}
}

var splitBatchTests = []struct {
	name     string
	proc     split
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"data",
		split{
			Options: splitOptions{
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
}

func TestSplitBatch(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()
	for _, test := range splitBatchTests {
		var capsules []config.Capsule
		for _, t := range test.test {
			capsule.SetData(t)
			capsules = append(capsules, capsule)
		}

		result, err := test.proc.Batch(ctx, capsules...)
		if err != nil {
			t.Error(err)
		}

		for i, res := range result {
			expected := test.expected[i]
			if !bytes.Equal(expected, res.Data()) {
				t.Errorf("expected %s, got %s", expected, string(res.Data()))
			}
		}
	}
}

func benchmarksplitBatch(b *testing.B, proc split, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Batch(ctx, capsules...)
	}
}

func BenchmarkSplitBatch(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range splitBatchTests {
		b.Run(test.name,
			func(b *testing.B) {
				var capsules []config.Capsule
				for _, t := range test.test {
					capsule.SetData(t)
					capsules = append(capsules, capsule)
				}
				benchmarksplitBatch(b, test.proc, capsules)
			},
		)
	}
}
