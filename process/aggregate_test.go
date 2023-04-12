package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var aggregateTests = []struct {
	name     string
	cfg      config.Config
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"single aggregate",
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"separator": `\n`,
				},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"separator": `\n`,
					"max_count": 3,
					"max_size":  50,
				},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"separator": `\n`,
					"max_count": 2,
				},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"set_key": "aggregate.-1",
			},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"set_key": "aggregate.-1",
				"options": map[string]interface{}{
					"max_count": 2,
				},
			},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"set_key": "aggregate.-1",
			},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"set_key": "aggregate.-1",
				"options": map[string]interface{}{
					"max_count": 2,
				},
			},
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
		config.Config{
			Type: "aggregate",
			Settings: map[string]interface{}{
				"set_key": "aggregate.-1",
				"options": map[string]interface{}{
					"key": "foo",
				},
			},
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
}

func TestAggregate(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range aggregateTests {
		t.Run(test.name, func(t *testing.T) {
			proc, err := newProcAggregate(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}
			var _ Batcher = proc

			var capsules []config.Capsule
			for _, t := range test.test {
				capsule.SetData(t)
				capsules = append(capsules, capsule)
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

func benchmarkAggregate(b *testing.B, batcher procAggregate, capsules []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = batcher.Batch(ctx, capsules...)
	}
}

func BenchmarkAggregate(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range aggregateTests {
		proc, err := newProcAggregate(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				var capsules []config.Capsule
				for _, t := range test.test {
					_ = capsule.SetData(t)
					capsules = append(capsules, capsule)
				}

				benchmarkAggregate(b, proc, capsules)
			},
		)
	}
}
