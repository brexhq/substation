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
	// strings
	{
		"delete",
		Delete{
			InputKey: "delete",
		},
		[]byte(`{"hello":"123","delete":"456"}`),
		[]byte(`{"hello":"123"}`),
		nil,
	},
}

func TestDelete(t *testing.T) {
	for _, test := range deleteTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
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
