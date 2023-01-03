package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var copyTests = []struct {
	name     string
	proc     procCopy
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		procCopy{
			process: process{
				Key:    "foo",
				SetKey: "baz",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","baz":"bar"}`),
		nil,
	},
	{
		"JSON unescape",
		procCopy{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
		},
		[]byte(`{"foo":"{\"bar\":\"baz\"}"`),
		[]byte(`{"foo":{"bar":"baz"}`),
		nil,
	},
	{
		"JSON unescape",
		procCopy{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
		},
		[]byte(`{"foo":"[\"bar\"]"}`),
		[]byte(`{"foo":["bar"]}`),
		nil,
	},
	{
		"from JSON",
		procCopy{
			process: process{
				Key: "foo",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`bar`),
		nil,
	},
	{
		"from JSON nested",
		procCopy{
			process: process{
				Key: "foo",
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON utf8",
		procCopy{
			process: process{
				SetKey: "bar",
			},
		},
		[]byte(`baz`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON base64",
		procCopy{
			process: process{
				SetKey: "bar",
			},
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[]byte(`{"bar":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
}

func TestCopy(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range copyTests {
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

func benchmarkCopy(b *testing.B, applier procCopy, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCopy(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range copyTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCopy(b, test.proc, capsule)
			},
		)
	}
}
