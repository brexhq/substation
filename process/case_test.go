package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var caseTests = []struct {
	name     string
	proc     procCase
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON lower",
		procCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procCaseOptions{
				Type: "lower",
			},
		},
		[]byte(`{"foo":"BAR"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON upper",
		procCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procCaseOptions{
				Type: "upper",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"BAR"}`),
		nil,
	},
	{
		"JSON snake",
		procCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procCaseOptions{
				Type: "snake",
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
		var _ Applier = test.proc
		var _ Batcher = test.proc

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

func benchmarkCase(b *testing.B, applier procCase, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCase(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range caseTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCase(b, test.proc, capsule)
			},
		)
	}
}
