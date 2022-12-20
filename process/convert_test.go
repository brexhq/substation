package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var convertTests = []struct {
	name     string
	proc     _convert
	test     []byte
	expected []byte
	err      error
}{
	{
		"bool true",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"foo":"true"}`),
		[]byte(`{"foo":true}`),
		nil,
	},
	{
		"bool false",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"foo":"false"}`),
		[]byte(`{"foo":false}`),
		nil,
	},
	{
		"int",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "int",
			},
		},
		[]byte(`{"foo":"-123"}`),
		[]byte(`{"foo":-123}`),
		nil,
	},
	{
		"float",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "float",
			},
		},
		[]byte(`{"foo":"123.456"}`),
		[]byte(`{"foo":123.456}`),
		nil,
	},
	{
		"uint",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "uint",
			},
		},
		[]byte(`{"foo":"123"}`),
		[]byte(`{"foo":123}`),
		nil,
	},
	{
		"string",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
				Type: "string",
			},
		},
		[]byte(`{"foo":123}`),
		[]byte(`{"foo":"123"}`),
		nil,
	},
	{
		"int",
		_convert{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _convertOptions{
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

func benchmarkConvert(b *testing.B, applicator _convert, test config.Capsule) {
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
