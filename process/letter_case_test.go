package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var letterCaseTests = []struct {
	name     string
	proc     letterCase
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON lower",
		letterCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: letterCaseOptions{
				Type: "lowercase",
			},
		},
		[]byte(`{"foo":"BAR"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON upper",
		letterCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: letterCaseOptions{
				Type: "uppercase",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"BAR"}`),
		nil,
	},
	{
		"JSON snake",
		letterCase{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: letterCaseOptions{
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

	for _, test := range letterCaseTests {
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

func benchmarkCase(b *testing.B, applicator letterCase, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkCase(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range letterCaseTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCase(b, test.proc, capsule)
			},
		)
	}
}
