package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var base64Tests = []struct {
	name     string
	proc     _base64
	test     []byte
	expected []byte
	err      error
}{
	{
		"data decode",
		_base64{
			Options: _base64Options{
				Direction: "from",
			},
		},
		[]byte(`YmFy`),
		[]byte(`bar`),
		nil,
	},
	{
		"data encode",
		_base64{
			Options: _base64Options{
				Direction: "to",
			},
		},
		[]byte(`bar`),
		[]byte(`YmFy`),
		nil,
	},
	{
		"JSON decode",
		_base64{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _base64Options{
				Direction: "from",
			},
		},
		[]byte(`{"foo":"YmFy"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestBase64(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range base64Tests {
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

func benchmarkbase64(b *testing.B, applier _base64, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkBase64(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range base64Tests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkbase64(b, test.proc, capsule)
			},
		)
	}
}
