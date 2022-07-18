package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
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
	for _, test := range caseTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkCaseByte(b *testing.B, byter Case, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkCaseByte(b *testing.B) {
	for _, test := range caseTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCaseByte(b, test.proc, test.test)
			},
		)
	}
}
