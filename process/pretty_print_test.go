package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var prettyPrintBatchTests = []struct {
	name     string
	proc     procPrettyPrint
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"from",
		procPrettyPrint{
			Options: procPrettyPrintOptions{
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
		procPrettyPrint{
			Options: procPrettyPrintOptions{
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
		procPrettyPrint{
			Options: procPrettyPrintOptions{
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
}

func TestPrettyPrintBatch(t *testing.T) {
	ctx := context.TODO()

	for _, test := range prettyPrintBatchTests {
		var _ Applier = test.proc
		var _ Batcher = test.proc

		var capsules []config.Capsule
		capsule := config.NewCapsule()
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
				t.Errorf("expected %s, got %s", expected, res.Data())
			}
		}
	}
}

func benchmarkPrettyPrintBatch(b *testing.B, batcher procPrettyPrint, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = batcher.Batch(ctx, capsules...)
	}
}

func BenchmarkPrettyPrintBatch(b *testing.B) {
	for _, test := range prettyPrintBatchTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsules := make([]config.Capsule, 1)
				for _, t := range test.test {
					capsule := config.NewCapsule()
					capsule.SetData(t)
					capsules = append(capsules, capsule)
				}

				benchmarkPrettyPrintBatch(b, test.proc, capsules)
			},
		)
	}
}

var prettyPrintTests = []struct {
	name     string
	proc     procPrettyPrint
	test     []byte
	expected []byte
	err      error
}{
	{
		"to",
		procPrettyPrint{
			Options: procPrettyPrintOptions{
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
}

func TestPrettyPrint(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range prettyPrintTests {
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

func benchmarkPrettyPrint(b *testing.B, proc procPrettyPrint, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Apply(ctx, capsule)
	}
}

func BenchmarkPrettyPrint(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range prettyPrintTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkPrettyPrint(b, test.proc, capsule)
			},
		)
	}
}
