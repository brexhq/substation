package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procSplit{}
	_ Batcher = procSplit{}
)

var splitTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		config.Config{
			Type: "split",
			Settings: map[string]interface{}{
				"key":     "split",
				"set_key": "split",
				"options": map[string]interface{}{
					"separator": ".",
				},
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcSplit(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
	}
}

func benchmarkSplit(b *testing.B, proc procSplit, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Apply(ctx, capsule)
	}
}

func BenchmarkSplit(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range splitTests {
		proc, err := newProcSplit(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkSplit(b, proc, capsule)
			},
		)
	}
}

var splitBatchTests = []struct {
	name     string
	cfg      config.Config
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"data",
		config.Config{
			Type: "split",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"separator": `\n`,
				},
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
		t.Run(test.name, func(t *testing.T) {
			var capsules []config.Capsule
			for _, t := range test.test {
				capsule.SetData(t)
				capsules = append(capsules, capsule)
			}

			proc, err := newProcSplit(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Batch(ctx, capsules...)
			if err != nil {
				t.Error(err)
			}

			for i, res := range result {
				expected := test.expected[i]
				if !bytes.Equal(expected, res.Data()) {
					t.Errorf("expected %s, got %s", expected, string(res.Data()))
				}
			}
		})
	}
}

func benchmarksplitBatch(b *testing.B, proc procSplit, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Batch(ctx, capsules...)
	}
}

func BenchmarkSplitBatch(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range splitBatchTests {
		proc, err := newProcSplit(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				var capsules []config.Capsule
				for _, t := range test.test {
					capsule.SetData(t)
					capsules = append(capsules, capsule)
				}
				benchmarksplitBatch(b, proc, capsules)
			},
		)
	}
}
