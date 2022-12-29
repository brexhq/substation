package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var hashTests = []struct {
	name     string
	proc     _hash
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON md5",
		_hash{
			process: process{
				Key:    "hash",
				SetKey: "hash",
			},
			Options: _hashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}`),
		nil,
	},
	{
		"JSON sha256",
		_hash{
			process: process{
				Key:    "hash",
				SetKey: "hash",
			},
			Options: _hashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"}`),
		nil,
	},
	{
		"JSON @this sha256",
		_hash{
			process: process{
				Key:    "@this",
				SetKey: "hash",
			},
			Options: _hashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[]byte(`{"hash":"6a65e3082b44c5da7fa58a5c335b2a95acab3a925c7fa0cfd5bd6779ee7c2374"}`),
		nil,
	},
	{
		"data md5",
		_hash{
			Options: _hashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`foo`),
		[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
		nil,
	},
	{
		"data sha256",
		_hash{
			Options: _hashOptions{
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

func benchmarkHash(b *testing.B, applier _hash, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
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
