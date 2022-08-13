package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var hashTests = []struct {
	name     string
	proc     Hash
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON md5",
		Hash{
			Options: HashOptions{
				Algorithm: "md5",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"37b51d194a7513e45b56f6524f2d51f2"}`),
		nil,
	},
	{
		"JSON sha256",
		Hash{
			Options: HashOptions{
				Algorithm: "sha256",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9"}`),
		nil,
	},
	{
		"JSON @this sha256",
		Hash{
			InputKey:  "@this",
			OutputKey: "foo",
			Options: HashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b"}`),
		nil,
	},
	{
		"data md5",
		Hash{
			Options: HashOptions{
				Algorithm: "md5",
			},
		},
		[]byte(`foo`),
		[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
		nil,
	},
	{
		"data sha256",
		Hash{
			Options: HashOptions{
				Algorithm: "sha256",
			},
		},
		[]byte(`foo`),
		[]byte(`2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae`),
		nil,
	},
	{
		"invalid settings",
		Hash{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestHash(t *testing.T) {
	ctx := context.TODO()
	for _, test := range hashTests {

		cap := config.NewCapsule()
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

func benchmarkHashCapByte(b *testing.B, applicator Hash, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkHashCapByte(b *testing.B) {
	for _, test := range hashTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkHashCapByte(b, test.proc, cap)
			},
		)
	}
}
