package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var hashTests = []struct {
	name     string
	proc     hash
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON md5",
		hash{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: hashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"37b51d194a7513e45b56f6524f2d51f2"}`),
		nil,
	},
	{
		"JSON sha256",
		hash{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: hashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9"}`),
		nil,
	},
	{
		"JSON @this sha256",
		hash{
			process: process{
				Key:    "@this",
				SetKey: "foo",
			},
			Options: hashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b"}`),
		nil,
	},
	{
		"data md5",
		hash{
			Options: hashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`foo`),
		[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
		nil,
	},
	{
		"data sha256",
		hash{
			Options: hashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`foo`),
		[]byte(`2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae`),
		nil,
	},
}

func TestHash(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range hashTests {
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

func benchmarkHash(b *testing.B, applicator hash, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkHash(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range hashTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkHash(b, test.proc, capsule)
			},
		)
	}
}
