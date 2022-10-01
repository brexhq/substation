package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var splitTests = []struct {
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
}

func TestSplit(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range splitTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.Data())
			t.Fail()
		}
	}
}

func benchmarkSplit(b *testing.B, proc Split, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		proc.Apply(ctx, cap)
	}
}

func BenchmarkSplit(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range splitTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkSplit(b, test.proc, cap)
			},
		)
	}
}

var splitBatchTests = []struct {
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
}

func TestSplitBatch(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range splitBatchTests {
		var caps []config.Capsule
		for _, t := range test.test {
			cap.SetData(t)
			caps = append(caps, cap)
		}

		result, err := test.proc.ApplyBatch(ctx, caps)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		for i, res := range result {
			expected := test.expected[i]
			if !bytes.Equal(expected, res.Data()) {
				t.Logf("expected %s, got %s", expected, string(res.Data()))
				t.Fail()
			}
		}
	}
}

func benchmarkSplitBatch(b *testing.B, proc Split, caps []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		proc.ApplyBatch(ctx, caps)
	}
}

func BenchmarkSplitBatch(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range splitBatchTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				var caps []config.Capsule
				for _, t := range test.test {
					cap.SetData(t)
					caps = append(caps, cap)
				}
				benchmarkSplitBatch(b, test.proc, caps)
			},
		)
	}
}
