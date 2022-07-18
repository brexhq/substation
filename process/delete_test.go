package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var deleteTests = []struct {
	name     string
	proc     Delete
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON",
		Delete{
			InputKey: "baz",
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	for _, test := range deleteTests {
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
func benchmarkDeleteByte(b *testing.B, byter Delete, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkDeleteByte(b *testing.B) {
	for _, test := range deleteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDeleteByte(b, test.proc, test.test)
			},
		)
	}
}
