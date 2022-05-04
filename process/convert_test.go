package process

import (
	"bytes"
	"context"
	"testing"
)

var convertTests = []struct {
	name     string
	proc     Convert
	test     []byte
	expected []byte
}{
	// strings
	{
		"json bool",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "bool",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":"true"}`),
		[]byte(`{"convert":true}`),
	},
	{
		"json bool",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "bool",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":"false"}`),
		[]byte(`{"convert":false}`),
	},
	{
		"json int",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "int",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":"-123"}`),
		[]byte(`{"convert":-123}`),
	},
	{
		"json float",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "float",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":"123.456"}`),
		[]byte(`{"convert":123.456}`),
	},
	{
		"json uint",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "uint",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":"123"}`),
		[]byte(`{"convert":123}`),
	},
	{
		"json string",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "string",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":123}`),
		[]byte(`{"convert":"123"}`),
	},
	{
		"json int",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "int",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":123.456}`),
		[]byte(`{"convert":123}`),
	},
	// array support
	{
		"json array bool",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "bool",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":["true","false"]}`),
		[]byte(`{"convert":[true,false]}`),
	},
	{
		"json array int",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "int",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":["-123","-456"]}`),
		[]byte(`{"convert":[-123,-456]}`),
	},
	{
		"json array float",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "float",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":["-123.456","123.456"]}`),
		[]byte(`{"convert":[-123.456,123.456]}`),
	},
	{
		"json array uint",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "uint",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":["123","456"]}`),
		[]byte(`{"convert":[123,456]}`),
	},
	{
		"json array string",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "string",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":[123,123.456]}`),
		[]byte(`{"convert":["123","123.456"]}`),
	},
	{
		"json array int",
		Convert{
			Input: Input{
				Key: "convert",
			},
			Options: ConvertOptions{
				Type: "int",
			},
			Output: Output{
				Key: "convert",
			},
		},
		[]byte(`{"convert":[123.456,1.2]}`),
		[]byte(`{"convert":[123,1]}`),
	},
}

func TestConvert(t *testing.T) {
	for _, test := range convertTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
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
