package condition

import (
	"testing"
)

var regexpTests = []struct {
	name      string
	inspector RegExp
	test      []byte
	expected  bool
}{
	{
		"pass",
		RegExp{
			Expression: "^Test",
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		RegExp{
			Expression: "^Test",
		},
		[]byte("-Test"),
		false,
	},
	{
		"!fail",
		RegExp{
			Negate:     true,
			Expression: "XYZ",
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		RegExp{
			Negate:     true,
			Expression: "ABC",
		},
		[]byte("ABC"),
		false,
	},
}

func TestRegExp(t *testing.T) {
	for _, rt := range regexpTests {
		check, _ := rt.inspector.Inspect(rt.test)

		if rt.expected != check {
			t.Logf("expected %v, got %v", rt.expected, check)
			t.Fail()
		}
	}
}

func benchmarkRegExpByte(b *testing.B, inspector RegExp, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkRegExpByte(b *testing.B) {
	for _, test := range regexpTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkRegExpByte(b, test.inspector, test.test)
			},
		)
	}
}
