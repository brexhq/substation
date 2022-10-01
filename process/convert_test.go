package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var convertTests = []struct {
	name     string
	proc     Convert
	test     []byte
	expected []byte
	err      error
}{
	{
		"bool true",
		Convert{
			Options: ConvertOptions{
				Type: "bool",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"true"}`),
		[]byte(`{"foo":true}`),
		nil,
	},
	{
		"bool false",
		Convert{
			Options: ConvertOptions{
				Type: "bool",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"false"}`),
		[]byte(`{"foo":false}`),
		nil,
	},
	{
		"int",
		Convert{
			Options: ConvertOptions{
				Type: "int",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"-123"}`),
		[]byte(`{"foo":-123}`),
		nil,
	},
	{
		"float",
		Convert{
			Options: ConvertOptions{
				Type: "float",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"123.456"}`),
		[]byte(`{"foo":123.456}`),
		nil,
	},
	{
		"uint",
		Convert{
			Options: ConvertOptions{
				Type: "uint",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"123"}`),
		[]byte(`{"foo":123}`),
		nil,
	},
	{
		"string",
		Convert{
			Options: ConvertOptions{
				Type: "string",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":123}`),
		[]byte(`{"foo":"123"}`),
		nil,
	},
	{
		"int",
		Convert{
			Options: ConvertOptions{
				Type: "int",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":123.456}`),
		[]byte(`{"foo":123}`),
		nil,
	},
}

func TestConvert(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range convertTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.Data())
			t.Fail()
		}
	}
}

func benchmarkConvert(b *testing.B, applicator Convert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkConvert(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range convertTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkConvert(b, test.proc, cap)
			},
		)
	}
}
