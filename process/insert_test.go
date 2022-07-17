package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var insertTests = []struct {
	name     string
	proc     Insert
	err      error
	test     []byte
	expected []byte
}{
	{
		"byte",
		Insert{
			Options: InsertOptions{
				Value: []byte{98, 97, 114},
			},
			OutputKey: "foo",
		},
		nil,
		[]byte{},
		[]byte(`{"foo":"bar"}`),
	},
	{
		"string",
		Insert{
			Options: InsertOptions{
				Value: "bar",
			},
			OutputKey: "foo",
		},
		nil,
		[]byte{},
		[]byte(`{"foo":"bar"}`),
	},
	{
		"int",
		Insert{
			Options: InsertOptions{
				Value: 10,
			},
			OutputKey: "int",
		},
		nil,
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","int":10}`),
	},
	{
		"string array",
		Insert{
			Options: InsertOptions{
				Value: []string{"bar", "baz", "qux"},
			},
			OutputKey: "array",
		},
		nil,
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","array":["bar","baz","qux"]}`),
	},
	{
		"map",
		Insert{
			Options: InsertOptions{
				Value: map[string]string{
					"baz": "qux",
				},
			},
			OutputKey: "map",
		},
		nil,
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","map":{"baz":"qux"}}`),
	},
	{
		"JSON",
		Insert{
			Options: InsertOptions{
				Value: `{"baz":"qux"}`,
			},
			OutputKey: "insert",
		},
		nil,
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","insert":{"baz":"qux"}}`),
	},
	{
		"zlib",
		Insert{
			Options: InsertOptions{
				Value: []byte{120, 156, 5, 192, 33, 13, 0, 0, 0, 128, 176, 182, 216, 247, 119, 44, 6, 2, 130, 1, 69},
			},
			OutputKey: "insert",
		},
		nil,
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","insert":"eJwFwCENAAAAgLC22Pd3LAYCggFF"}`),
	},
	{
		"missing required options",
		Insert{},
		ProcessorInvalidSettings,
		[]byte{},
		[]byte{},
	},
}

func TestInsert(t *testing.T) {
	ctx := context.TODO()
	for _, test := range insertTests {
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

func benchmarkInsertByte(b *testing.B, byter Insert, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkInsertByte(b *testing.B) {
	for _, test := range insertTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkInsertByte(b, test.proc, test.test)
			},
		)
	}
}
