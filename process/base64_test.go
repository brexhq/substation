package process

import (
	"bytes"
	"context"
	"testing"
)

var base64Tests = []struct {
	name     string
	proc     Base64
	test     []byte
	expected []byte
}{
	// decode std base64
	{
		"decode data std",
		Base64{
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "std",
			},
		},
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB+`),
		[]byte(`abc123!?$*&()'-=@~`),
	},
	// decode url base64
	{
		"decode data url",
		Base64{
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "url",
			},
		},
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB-`),
		[]byte(`abc123!?$*&()'-=@~`),
	},
	// encode std base64
	{
		"encode data std",
		Base64{
			Options: Base64Options{
				Direction: "to",
				Alphabet:  "std",
			},
		},
		[]byte(`abc123!?$*&()'-=@~`),
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB+`),
	},
	// encode url base64
	{
		"encode data url",
		Base64{
			Options: Base64Options{
				Direction: "to",
				Alphabet:  "url",
			},
		},
		[]byte(`abc123!?$*&()'-=@~`),
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB-`),
	},
	// decode std base64 from input to output
	{
		"json std",
		Base64{
			Input: Input{
				Key: "base64",
			},
			Output: Output{
				Key: "base64",
			},
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "std",
			},
		},
		[]byte(`{"base64":"YWJjMTIzIT8kKiYoKSctPUB+"}`),
		[]byte(`{"base64":"abc123!?$*&()'-=@~"}`),
	},
	// decode array of std base64 from input to output
	{
		"json array std",
		Base64{
			Input: Input{
				Key: "base64",
			},
			Output: Output{
				Key: "base64",
			},
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "std",
			},
		},
		[]byte(`{"base64":["aGVsbG8=","d29ybGQ="]}`),
		[]byte(`{"base64":["hello","world"]}`),
	},
}

func TestBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range base64Tests {
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

func benchmarkBase64Byte(b *testing.B, byter Base64, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkBase64Byte(b *testing.B) {
	for _, test := range base64Tests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkBase64Byte(b, test.proc, test.test)
			},
		)
	}
}
