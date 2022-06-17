package process

import (
	"bytes"
	"context"
	"testing"
)

var hashTests = []struct {
	name     string
	proc     Hash
	test     []byte
	expected []byte
}{
	{
		"json md5",
		Hash{
			Input:  "hash",
			Output: "hash",
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}`),
	},
	{
		"json sha256",
		Hash{
			Input:  "hash",
			Output: "hash",
			Options: HashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"}`),
	},
	{
		"json @this md5",
		Hash{
			Input:  "@this",
			Output: "hash",
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"92568430636b469705c89466dfd39646"}`),
	},
	{
		"json array md5",
		Hash{
			Input:  "hash",
			Output: "hash",
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`{"hash":["foo","bar"]}`),
		[]byte(`{"hash":["acbd18db4cc2f85cedef654fccc4a4d8","37b51d194a7513e45b56f6524f2d51f2"]}`),
	},
	{
		"data",
		Hash{
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`foo`),
		[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
	},
}

func TestHash(t *testing.T) {
	ctx := context.TODO()
	for _, test := range hashTests {
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
