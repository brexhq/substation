package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
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
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
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

func benchmarkGroupByte(b *testing.B, byter Group, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkGroupByte(b *testing.B) {
	for _, test := range groupTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkGroupByte(b, test.proc, test.test)
			},
		)
	}
}
