package process

import (
	"bytes"
	"context"
	"testing"
)

var flattenTests = []struct {
	name     string
	proc     Flatten
	test     []byte
	expected []byte
}{
	{
		"json",
		Flatten{
			InputKey:  "flatten",
			OutputKey: "flatten",
		},
		[]byte(`{"flatten":["foo",["bar"]]}`),
		[]byte(`{"flatten":["foo","bar"]}`),
	},
	{
		"json deep flatten",
		Flatten{
			InputKey:  "flatten",
			OutputKey: "flatten",
			Options: FlattenOptions{
				Deep: true,
			},
		},
		[]byte(`{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}`),
		[]byte(`{"flatten":["foo","bar","baz"]}`),
	},
}

func TestFlatten(t *testing.T) {
	for _, test := range flattenTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
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
