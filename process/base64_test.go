package process

import (
	"bytes"
	"context"
	"testing"
)

var base64Tests = []struct {
	proc     Base64
	test     []byte
	expected []byte
}{
	// decode std base64
	{
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
		Base64{
			Options: Base64Options{
				Direction: "to",
				Alphabet:  "url",
			},
		},
		[]byte(`abc123!?$*&()'-=@~`),
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB-`),
	},
	// decode std base64 from input
	{
		Base64{
			Input: Input{
				Key: "base64",
			},
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "std",
			},
		},
		[]byte(`{"base64":"YWJjMTIzIT8kKiYoKSctPUB+"}`),
		[]byte(`abc123!?$*&()'-=@~`),
	},
	// decode std base64 to output
	{
		Base64{
			Output: Output{
				Key: "base64",
			},
			Options: Base64Options{
				Direction: "from",
				Alphabet:  "std",
			},
		},
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB+`),
		[]byte(`{"base64":"abc123!?$*&()'-=@~"}`),
	},
	// decode std base64 from input to output
	{
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
