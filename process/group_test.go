package process

import (
	"bytes"
	"context"
	"testing"
)

var groupTests = []struct {
	name string
	proc Group
	test []byte
	// the order of the grouped output is inconsistent, so we check for a match anywhere in this slice
	expected [][]byte
}{
	{
		"tuples",
		Group{
			Input: Inputs{
				Keys: []string{"g1", "g2"},
			},
			Output: Output{
				Key: "group",
			},
		},
		[]byte(`{"g1":["foo","bar"],"g2":[123,456]}`),
		[][]byte{
			[]byte(`{"g1":["foo","bar"],"g2":[123,456],"group":[["foo",123],["bar",456]]}`),
			[]byte(`{"g1":["foo","bar"],"g2":[123,456],"group":[["bar",456],["foo",123]]}`),
		},
	},
	{
		"objects",
		Group{
			Input: Inputs{
				Keys: []string{"g1", "g2"},
			},
			Options: GroupOptions{
				Keys: []string{"name.test", "size"},
			},
			Output: Output{
				Key: "group",
			},
		},
		[]byte(`{"g1":["foo","bar"],"g2":[123,456]}`),
		[][]byte{
			[]byte(`{"g1":["foo","bar"],"g2":[123,456],"group":[{"name":{"test":"foo"},"size":123},{"name":{"test":"bar"},"size":456}]}`),
			[]byte(`{"g1":["foo","bar"],"g2":[123,456],"group":[{"name":{"test":"bar"},"size":456},{"name":{"test":"foo"},"size":123}]}`),
		},
	},
}

func TestGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range groupTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		pass := false
		for _, x := range test.expected {
			if c := bytes.Compare(res, x); c == 0 {
				pass = true
			}
		}

		if !pass {
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
