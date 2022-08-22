package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var caseTests = []struct {
	name     string
	proc     Case
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON lower",
		Case{
			Options: CaseOptions{
				Case: "lower",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"BAR"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON upper",
		Case{
			Options: CaseOptions{
				Case: "upper",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"BAR"}`),
		nil,
	},
	{
		"JSON snake",
		Case{
			InputKey:  "foo",
			OutputKey: "foo",
			Options: CaseOptions{
				Case: "snake",
			},
		},
		[]byte(`{"foo":"AbC"})`),
		[]byte(`{"foo":"ab_c"})`),
		nil,
	},
	{
		"invalid settings",
		Case{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestCase(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range caseTests {
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkCase(b *testing.B, applicator Case, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkCase(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range caseTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkCase(b, test.proc, cap)
			},
		)
	}
}
