package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
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
}

func TestInsert(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range insertTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkInsert(b *testing.B, applicator Insert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkInsert(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range insertTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkInsert(b, test.proc, capsule)
			},
		)
	}
}
