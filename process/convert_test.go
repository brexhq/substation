package process

import (
	"bytes"
	"context"
	"testing"
)

func TestConvert(t *testing.T) {
	var tests = []struct {
		proc     Convert
		test     []byte
		expected []byte
	}{
		// strings
		{
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

	for _, test := range tests {
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
