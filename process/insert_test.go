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
	test     []byte
	expected []byte
	err      error
}{
	{
		"byte",
		Insert{
			Options: InsertOptions{
				Value: []byte{98, 97, 114},
			},
			OutputKey: "foo",
		},
		[]byte{},
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"string",
		Insert{
			Options: InsertOptions{
				Value: "bar",
			},
			OutputKey: "foo",
		},
		[]byte{},
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"int",
		Insert{
			Options: InsertOptions{
				Value: 10,
			},
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":10}`),
		nil,
	},
	{
		"string array",
		Insert{
			Options: InsertOptions{
				Value: []string{"bar", "baz", "qux"},
			},
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":["bar","baz","qux"]}`),
		nil,
	},
	{
		"map",
		Insert{
			Options: InsertOptions{
				Value: map[string]string{
					"baz": "qux",
				},
			},
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
	},
	{
		"JSON",
		Insert{
			Options: InsertOptions{
				Value: `{"baz":"qux"}`,
			},
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
	},
	{
		"zlib",
		Insert{
			Options: InsertOptions{
				Value: []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
			},
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
	{
		"invalid settings",
		Insert{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestInsert(t *testing.T) {
	ctx := context.TODO()
	for _, test := range insertTests {
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
