package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var replaceTests = []struct {
	name     string
	proc     Replace
	err      error
	test     []byte
	expected []byte
}{
	{
		"json",
		Replace{
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
			InputKey:  "replace",
			OutputKey: "replace",
		},
		nil,
		[]byte(`{"replace":"bar"}`),
		[]byte(`{"replace":"baz"}`),
	},
	{
		"data",
		Replace{
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
		},
		nil,
		[]byte(`bar`),
		[]byte(`baz`),
	},
	{
		"missing required options",
		Replace{
			Options: ReplaceOptions{},
		},
		nil,
		[]byte{},
		[]byte{},
	},
}

func TestReplace(t *testing.T) {
	ctx := context.TODO()
	for _, test := range replaceTests {
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

func benchmarkReplaceByte(b *testing.B, byter Replace, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkReplaceByte(b *testing.B) {
	for _, test := range replaceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkReplaceByte(b, test.proc, test.test)
			},
		)
	}
}
