package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var base64Tests = []struct {
	name     string
	proc     Base64
	test     []byte
	expected []byte
	err      error
}{
	{
		"data decode",
		Base64{
			Options: Base64Options{
				Direction: "from",
			},
		},
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB+`),
		[]byte(`abc123!?$*&()'-=@~`),
		nil,
	},
	{
		"data encode",
		Base64{
			Options: Base64Options{
				Direction: "to",
			},
		},
		[]byte(`abc123!?$*&()'-=@~`),
		[]byte(`YWJjMTIzIT8kKiYoKSctPUB+`),
		nil,
	},
	{
		"JSON decode",
		Base64{
			Options: Base64Options{
				Direction: "from",
			},
			InputKey:  "base64",
			OutputKey: "base64",
		},
		[]byte(`{"base64":"YWJjMTIzIT8kKiYoKSctPUB+"}`),
		[]byte(`{"base64":"abc123!?$*&()'-=@~"}`),
		nil,
	},
	{
		"invalid settings",
		Base64{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
	{
		"invalid direction",
		Base64{
			Options: Base64Options{
				Direction: "foo",
			},
			InputKey:  "base64",
			OutputKey: "base64",
		},
		[]byte(`{"base64":"H4sIAMSJy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[]byte(``),
		ProcessorInvalidDirection,
	},
	{
		"JSON binary",
		Base64{
			Options: Base64Options{
				Direction: "from",
			},
			InputKey:  "base64",
			OutputKey: "base64",
		},
		[]byte(`{"base64":"H4sIAMSJy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA"}`),
		[]byte(``),
		Base64JSONDecodedBinary,
	},
}

func TestBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range base64Tests {
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
