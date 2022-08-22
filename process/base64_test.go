package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
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
		[]byte(`YmFy`),
		[]byte(`bar`),
		nil,
	},
	{
		"data encode",
		Base64{
			Options: Base64Options{
				Direction: "to",
			},
		},
		[]byte(`bar`),
		[]byte(`YmFy`),
		nil,
	},
	{
		"JSON decode",
		Base64{
			Options: Base64Options{
				Direction: "from",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"YmFy"}`),
		[]byte(`{"foo":"bar"}`),
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
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"YmFy"}`),
		[]byte(``),
		Base64InvalidDirection,
	},
	{
		"JSON binary",
		Base64{
			Options: Base64Options{
				Direction: "from",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		[]byte(``),
		Base64JSONDecodedBinary,
	},
}

func TestBase64(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range base64Tests {
		cap.SetData(test.test)

		res, err := test.proc.Apply(ctx, cap)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res.GetData(), test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res.GetData())
			t.Fail()
		}
	}
}

func benchmarkBase64(b *testing.B, applicator Base64, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkBase64(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range base64Tests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkBase64(b, test.proc, cap)
			},
		)
	}
}
