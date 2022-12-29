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
	proc     _replace
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		_replace{
			process: process{
				Key:    "replace",
				SetKey: "replace",
			},
			Options: _replaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`{"replace":"bar"}`),
		[]byte(`{"replace":"baz"}`),
		nil,
	},
	{
		"json delete",
		_replace{
			process: process{
				Key:    "replace",
				SetKey: "replace",
			},
			Options: _replaceOptions{
				Old: "z",
				New: "",
			},
		},
		[]byte(`{"replace":"fizz"}`),
		[]byte(`{"replace":"fi"}`),
		nil,
	},
	{
		"data",
		_replace{
			Options: _replaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`bar`),
		[]byte(`baz`),
		nil,
	},
	{
		"data delete",
		_replace{
			Options: _replaceOptions{
				Old: "r",
				New: "",
			},
		},
		[]byte(`bar`),
		[]byte(`ba`),
		nil,
	},
	{
		"data",
		_replace{
			Options: _replaceOptions{
				New: "z",
			},
		},
		[]byte(`bar`),
		[]byte(`baz`),
		errMissingRequiredOptions,
	},
}

func TestReplace(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range replaceTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			if errors.Is(err, test.err) {
				continue
			}
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkReplace(b *testing.B, applier _replace, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkReplace(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range replaceTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkReplace(b, test.proc, capsule)
			},
		)
	}
}
