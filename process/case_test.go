package process

import (
	"bytes"
	"context"
	"testing"
)

var caseTests = []struct {
	name     string
	proc     Case
	test     []byte
	expected []byte
}{
	{
		"json lower",
		Case{
			Input:  "case",
			Output: "case",
			Options: CaseOptions{
				Case: "lower",
			},
		},
		[]byte(`{"case":"ABC"}`),
		[]byte(`{"case":"abc"}`),
	},
	{
		"json upper",
		Case{
			Input:  "case",
			Output: "case",
			Options: CaseOptions{
				Case: "upper",
			},
		},
		[]byte(`{"case":"abc"}`),
		[]byte(`{"case":"ABC"}`),
	},
	{
		"json snake",
		Case{
			Input:  "case",
			Output: "case",
			Options: CaseOptions{
				Case: "snake",
			},
		},
		[]byte(`{"case":"AbC"})`),
		[]byte(`{"case":"ab_c"})`),
	},
	// array support
	{
		"json array lower",
		Case{
			Input:  "case",
			Output: "case",
			Options: CaseOptions{
				Case: "lower",
			},
		},
		[]byte(`{"case":["ABC","DEF"]}`),
		[]byte(`{"case":["abc","def"]}`),
	},
	{
		"data",
		Case{
			Options: CaseOptions{
				Case: "upper",
			},
		},
		[]byte(`foo`),
		[]byte(`FOO`),
	},
}

func TestCase(t *testing.T) {
	for _, test := range caseTests {
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

func benchmarkCaseByte(b *testing.B, byter Case, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkCaseByte(b *testing.B) {
	for _, test := range caseTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCaseByte(b, test.proc, test.test)
			},
		)
	}
}
