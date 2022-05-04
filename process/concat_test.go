package process

import (
	"bytes"
	"context"
	"testing"
)

var concatTests = []struct {
	name     string
	proc     Concat
	test     []byte
	expected []byte
}{
	{
		"json",
		Concat{
			Input: Inputs{
				Keys: []string{"c1", "c2"},
			},
			Options: ConcatOptions{
				Separator: ".",
			},
			Output: Output{
				Key: "c3",
			},
		},
		[]byte(`{"c1":"foo","c2":"bar"}`),
		[]byte(`{"c1":"foo","c2":"bar","c3":"foo.bar"}`),
	},
}

func TestConcat(t *testing.T) {
	for _, test := range concatTests {
		ctx := context.TODO()
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

func TestConcatArray(t *testing.T) {
	var tests = []struct {
		proc Concat
		test []byte
		// the order of the concat output when used on arrays is inconsistent, so we check for a match anywhere in this slice
		expected [][]byte
	}{
		{
			Concat{
				Input: Inputs{
					Keys: []string{"c1", "c2"},
				},
				Options: ConcatOptions{
					Separator: ".",
				},
				Output: Output{
					Key: "c3",
				},
			},
			[]byte(`{"c1":["foo","baz"],"c2":["bar","qux"]}`),
			[][]byte{
				[]byte(`{"c1":["foo","baz"],"c2":["bar","qux"],"c3":["foo.bar","baz.qux"]}`),
				[]byte(`{"c1":["foo","baz"],"c2":["bar","qux"],"c3":["baz.qux","foo.bar"]}`),
			},
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
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

func benchmarkConcatByte(b *testing.B, byter Concat, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkConcatByte(b *testing.B) {
	for _, test := range concatTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkConcatByte(b, test.proc, test.test)
			},
		)
	}
}
