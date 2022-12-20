package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var insertTests = []struct {
	name     string
	proc     _insert
	test     []byte
	expected []byte
	err      error
}{
	{
		"byte",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
				Value: []byte{102, 111, 111},
			},
		},
		[]byte{},
		[]byte(`{"insert":"foo"}`),
		nil,
	},
	{
		"string",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
				Value: "foo",
			},
		},
		[]byte{},
		[]byte(`{"insert":"foo"}`),
		nil,
	},
	{
		"int",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
				Value: 10,
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":10}`),
		nil,
	},
	{
		"string array",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
				Value: []string{"bar", "baz"},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":["bar","baz"]}`),
		nil,
	},
	{
		"map",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
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
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
				Value: `{"bar":"baz"}`,
			},
		},
		[]byte(`{"insert":"bar"}`),
		[]byte(`{"insert":{"bar":"baz"}}`),
		nil,
	},
	{
		"zlib",
		_insert{
			process: process{
				SetKey: "insert",
			},
			Options: _insertOptions{
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

func benchmarkInsert(b *testing.B, applicator _insert, test config.Capsule) {
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
