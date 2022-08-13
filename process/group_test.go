package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var groupTests = []struct {
	name     string
	proc     Group
	test     []byte
	expected []byte
	err      error
}{
	{
		"tuples",
		Group{
			InputKey:  "group",
			OutputKey: "group",
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[["foo",123],["bar",456]]}`),
		nil,
	},
	{
		"objects",
		Group{
			Options: GroupOptions{
				Keys: []string{"name.test", "size"},
			},
			InputKey:  "group",
			OutputKey: "group",
		},
		[]byte(`{"group":[["foo","bar"],[123,456]]}`),
		[]byte(`{"group":[{"name":{"test":"foo"},"size":123},{"name":{"test":"bar"},"size":456}]}`),
		nil,
	},
	{
		"invalid settings",
		Group{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range groupTests {

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

func benchmarkGroupCapByte(b *testing.B, applicator Group, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkGroupCapByte(b *testing.B) {
	for _, test := range groupTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap := config.NewCapsule()
				cap.SetData(test.test)
				benchmarkGroupCapByte(b, test.proc, cap)
			},
		)
	}
}
