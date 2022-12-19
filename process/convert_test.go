package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var convertTests = []struct {
	name     string
	proc     convert
	test     []byte
	expected []byte
	err      error
}{
	{
		"bool true",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"foo":"true"}`),
		[]byte(`{"foo":true}`),
		nil,
	},
	{
		"bool false",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"foo":"false"}`),
		[]byte(`{"foo":false}`),
		nil,
	},
	{
		"int",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "int",
			},
		},
		[]byte(`{"foo":"-123"}`),
		[]byte(`{"foo":-123}`),
		nil,
	},
	{
		"float",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "float",
			},
		},
		[]byte(`{"foo":"123.456"}`),
		[]byte(`{"foo":123.456}`),
		nil,
	},
	{
		"uint",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "uint",
			},
		},
		[]byte(`{"foo":"123"}`),
		[]byte(`{"foo":123}`),
		nil,
	},
	{
		"string",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "string",
			},
		},
		[]byte(`{"foo":123}`),
		[]byte(`{"foo":"123"}`),
		nil,
	},
	{
		"int",
		convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: convertOptions{
				Type: "int",
			},
		},
		[]byte(`{"foo":123.456}`),
		[]byte(`{"foo":123}`),
		nil,
	},
}

func TestConvert(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range convertTests {
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

func benchmarkConvert(b *testing.B, applicator convert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkConvert(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range convertTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkConvert(b, test.proc, capsule)
			},
		)
	}
}
