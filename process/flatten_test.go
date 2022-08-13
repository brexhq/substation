package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var flattenTests = []struct {
	name     string
	proc     Flatten
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		Flatten{
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[]byte(`{"flatten":["foo","bar"]}`),
		nil,
	},
	{
		"json deep flatten",
		Flatten{
			Options: FlattenOptions{
				Deep: true,
			},
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
		nil,
	},
	{
		"invalid settings",
		Flatten{
			Options: FlattenOptions{
				Deep: true,
			},
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
		ProcessorInvalidSettings,
	},
}

func TestFlatten(t *testing.T) {
	ctx := context.TODO()
	for _, test := range flattenTests {

		cap := config.NewCapsule()
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

func benchmarkFlattenCapByte(b *testing.B, applicator Flatten, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkFlattenCapByte(b *testing.B) {
	for _, test := range flattenTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkFlattenCapByte(b, test.proc, cap)
			},
		)
	}
}
