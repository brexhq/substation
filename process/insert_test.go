package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var insertTests = []struct {
	name     string
	proc     insert
	test     []byte
	expected []byte
	err      error
}{
	{
		"byte",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: []byte{98, 97, 114},
			},
		},
		[]byte{},
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"string",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: "bar",
			},
		},
		[]byte{},
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"int",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: 10,
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":10}`),
		nil,
	},
	{
		"string array",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: []string{"bar", "baz", "qux"},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":["bar","baz","qux"]}`),
		nil,
	},
	{
		"map",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: map[string]string{
					"baz": "qux",
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
	},
	{
		"JSON",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: `{"baz":"qux"}`,
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":{"baz":"qux"}}`),
		nil,
	},
	{
		"zlib",
		insert{
			process: process{
				SetKey: "foo",
			},
			Options: insertOptions{
				Value: []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
			},
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

func benchmarkInsert(b *testing.B, applicator insert, test config.Capsule) {
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
