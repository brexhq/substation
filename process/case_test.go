package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procCase{}
	_ Batcher = procCase{}
)

var caseTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON lower",
		config.Config{
			Type: "case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "lower",
				},
			},
		},
		[]byte(`{"foo":"BAR"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON upper",
		config.Config{
			Type: "case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "upper",
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"BAR"}`),
		nil,
	},
	{
		"JSON snake",
		config.Config{
			Type: "case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "snake",
				},
			},
		},
		[]byte(`{"foo":"AbC"})`),
		[]byte(`{"foo":"ab_c"})`),
		nil,
	},
}

func TestCase(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range caseTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcCase(ctx, test.cfg)
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

func benchmarkCase(b *testing.B, applier procCase, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCase(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range caseTests {
		proc, err := newProcCase(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCase(b, proc, capsule)
			},
		)
	}
}
