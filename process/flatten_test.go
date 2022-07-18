package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
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

func benchmarkFlattenByte(b *testing.B, byter Flatten, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkFlattenByte(b *testing.B) {
	for _, test := range flattenTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkFlattenByte(b, test.proc, test.test)
			},
		)
	}
}
