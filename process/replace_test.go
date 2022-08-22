package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var replaceTests = []struct {
	name     string
	proc     Replace
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		Replace{
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"baz"}`),
		nil,
	},
	{
		"data",
		Replace{
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`bar`),
		[]byte(`baz`),
		nil,
	},
	{
		"invalid settings",
		Replace{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestReplace(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range replaceTests {
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

func benchmarkReplace(b *testing.B, applicator Replace, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkReplace(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range replaceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkReplace(b, test.proc, cap)
			},
		)
	}
}
