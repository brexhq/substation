package process

import (
	"bytes"
	"context"
	"testing"
)

var insertTests = []struct {
	name     string
	proc     Insert
	test     []byte
	expected []byte
}{
	{
		"byte",
		Insert{
			Options: InsertOptions{
				Value: []byte{98, 97, 114},
			},
			Output: "foo",
		},
		[]byte(``),
		[]byte(`{"foo":"bar"}`),
	},
	{
		"string",
		Insert{
			Options: InsertOptions{
				Value: "bar",
			},
			Output: "foo",
		},
		[]byte(``),
		[]byte(`{"foo":"bar"}`),
	},
	{
		"int",
		Insert{
			Options: InsertOptions{
				Value: 10,
			},
			Output: "int",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","int":10}`),
	},
	{
		"string array",
		Insert{
			Options: InsertOptions{
				Value: []string{"bar", "baz", "qux"},
			},
			Output: "array",
		},
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
			Output: "map",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","map":{"baz":"qux"}}`),
	},
}

func TestInsert(t *testing.T) {
	for _, test := range insertTests {
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
