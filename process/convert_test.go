package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
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
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"convert":"true"}`),
		[]byte(`{"convert":true}`),
		nil,
	},
	{
		"bool false",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "bool",
			},
		},
		[]byte(`{"convert":"false"}`),
		[]byte(`{"convert":false}`),
		nil,
	},
	{
		"int",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "int",
			},
		},
		[]byte(`{"convert":"-123"}`),
		[]byte(`{"convert":-123}`),
		nil,
	},
	{
		"float",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "float",
			},
		},
		[]byte(`{"convert":"123.456"}`),
		[]byte(`{"convert":123.456}`),
		nil,
	},
	{
		"uint",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "uint",
			},
		},
		[]byte(`{"convert":"123"}`),
		[]byte(`{"convert":123}`),
		nil,
	},
	{
		"string",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "string",
			},
		},
		[]byte(`{"convert":123}`),
		[]byte(`{"convert":"123"}`),
		nil,
	},
	{
		"int",
		Convert{
			InputKey:  "convert",
			OutputKey: "convert",
			Options: ConvertOptions{
				Type: "int",
			},
		},
		[]byte(`{"convert":123.456}`),
		[]byte(`{"convert":123}`),
		nil,
	},
}

func TestConvert(t *testing.T) {
	for _, test := range convertTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
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

func benchmarkConvertByte(b *testing.B, byter Convert, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkConvertByte(b *testing.B) {
	for _, test := range convertTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkConvertByte(b, test.proc, test.test)
			},
		)
	}
}
