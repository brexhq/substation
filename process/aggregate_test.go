package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var aggregateTests = []struct {
	name     string
	proc     Aggregate
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"single aggregate",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`),
		},
		nil,
	},
	// identical to the single buffer test, but improves performance due to lowered count and size limits
	{
		"single aggregate with better performance",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
				MaxCount:  3,
				MaxSize:   50,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`),
		},
		nil,
	},
	{
		"multiple aggregates",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
				MaxCount:  2,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}\n{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		nil,
	},
	{
		"single JSON array aggregate",
		Aggregate{
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"baz":"qux"},{"quux":"corge"}]}`),
			[]byte(`{"aggregate":[{"fofo":"bar"},{"baz":"qux"},{"quux":"corge"}]}`),
		},
		nil,
	},
	{
		"multiple JSON array aggregates",
		Aggregate{
			Options: AggregateOptions{
				MaxCount: 2,
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
			[]byte(`{"quux":"corge"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"baz":"qux"}]}`),
			[]byte(`{"aggregate":[{"quux":"corge"}]}`),
		},
		nil,
	},
	{
		"single JSON array aggregate",
		Aggregate{
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`foo`),
			[]byte(`bar`),
			[]byte(`baz`),
			[]byte(`qux`),
			[]byte(`quux`),
			[]byte(`corge`),
		},
		[][]byte{
			[]byte(`{"aggregate":["foo","bar","baz","qux","quux","corge"]}`),
		},
		nil,
	},
	{
		"multiple JSON array aggregates",
		Aggregate{
			Options: AggregateOptions{
				MaxCount: 2,
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`foo`),
			[]byte(`bar`),
			[]byte(`baz`),
			[]byte(`qux`),
			[]byte(`quux`),
			[]byte(`corge`),
		},
		[][]byte{
			[]byte(`{"aggregate":["foo","bar"]}`),
			[]byte(`{"aggregate":["baz","qux"]}`),
			[]byte(`{"aggregate":["quux","corge"]}`),
		},
		nil,
	},
	{
		"JSON key aggregates",
		Aggregate{
			Options: AggregateOptions{
				AggregateKey: "foo",
			},
			OutputKey: "aggregate.-1",
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"qux"}`),
			[]byte(`{"foo":"bar"}`),
		},
		[][]byte{
			[]byte(`{"aggregate":[{"foo":"bar"},{"foo":"bar"},{"foo":"bar"}]}`),
			[]byte(`{"aggregate":[{"foo":"baz"}]}`),
			[]byte(`{"aggregate":[{"foo":"qux"}]}`),
		},
		nil,
	},
	// results in error AggregateBufferSizeLimit due to MaxSize limit of 1 byte
	{
		"buffer size limit",
		Aggregate{
			Options: AggregateOptions{
				Separator: `\n`,
				MaxSize:   1,
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"baz":"qux"}`),
		},
		[][]byte{},
		AggregateBufferSizeLimit,
	},
}

func TestAggregate(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range aggregateTests {
		var caps []config.Capsule
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
				t.Logf("expected %s, got %s", expected, string(processed.GetData()))
				t.Fail()
			}
		}
	}
}

func benchmarkAggregate(b *testing.B, batcher Aggregate, caps []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		batcher.ApplyBatch(ctx, caps)
	}
}

func BenchmarkAggregate(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range aggregateTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				var caps []config.Capsule
				for _, t := range test.test {
					cap.SetData(t)
					caps = append(caps, cap)
				}

				benchmarkAggregate(b, test.proc, caps)
			},
		)
	}
}
