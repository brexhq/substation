package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var insertTests = []struct {
	name     string
	proc     procInsert
	test     []byte
	expected []byte
	err      error
}{
	{
		"byte",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: []byte{102, 111, 111},
			},
		},
		[]byte{},
		[]byte(`{"insert":"foo"}`),
		nil,
	},
	{
		"string",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: "foo",
			},
		},
		[]byte{},
		[]byte(`{"insert":"foo"}`),
		nil,
	},
	{
		"int",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: 10,
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":10}`),
		nil,
	},
	{
		"string array",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: []string{"bar", "baz"},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":["bar","baz"]}`),
		nil,
	},
	{
		"map",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: map[string]string{
					"bar": "baz",
				},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":{"bar":"baz"}}`),
		nil,
	},
	{
		"JSON",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: `{"bar":"baz"}`,
			},
		},
		[]byte(`{"insert":"bar"}`),
		[]byte(`{"insert":{"bar":"baz"}}`),
		nil,
	},
	{
		"zlib",
		procInsert{
			process: process{
				SetKey: "insert",
			},
			Options: procInsertOptions{
				Value: []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
			},
		},
		[]byte(`{"insert":"bar"}`),
		[]byte(`{"insert":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
}

func TestInsert(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range insertTests {
		var _ Applier = test.proc
		var _ Batcher = test.proc

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

func benchmarkInsert(b *testing.B, applier procInsert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
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
