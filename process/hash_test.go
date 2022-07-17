package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var hashTests = []struct {
	name     string
	proc     Hash
	err      error
	test     []byte
	expected []byte
}{
	{
		"json md5",
		Hash{
			InputKey:  "hash",
			OutputKey: "hash",
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		nil,
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}`),
	},
	{
		"json sha256",
		Hash{
			InputKey:  "hash",
			OutputKey: "hash",
			Options: HashOptions{
				Algorithm: "sha256",
			},
		},
		nil,
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"}`),
	},
	{
		"json @this md5",
		Hash{
			InputKey:  "@this",
			OutputKey: "hash",
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		nil,
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"92568430636b469705c89466dfd39646"}`),
	},
	{
		"data",
		Hash{
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		nil,
		[]byte(`foo`),
		[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
	},
	{
		"missing required options",
		Hash{},
		ProcessorInvalidSettings,
		[]byte{},
		[]byte{},
	},
	{
		"unsupported algorithm",
		Hash{
			Options: HashOptions{
				Algorithm: "foo",
			},
		},
		HashUnsupportedAlgorithm,
		[]byte(`foo`),
		[]byte{},
	},
}

func TestHash(t *testing.T) {
	ctx := context.TODO()
	for _, test := range hashTests {
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

func benchmarkHashByte(b *testing.B, byter Hash, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkHashByte(b *testing.B) {
	for _, test := range hashTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkHashByte(b, test.proc, test.test)
			},
		)
	}
}
