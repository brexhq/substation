package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
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
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range regexpTests {
		cap.SetData(test.test)
		check, _ := test.inspector.Inspect(ctx, cap)

		if test.expected != check {
			t.Logf("expected %v, got %v", test.expected, check)
			t.Fail()
		}
	}
}

func benchmarkRegExpByte(b *testing.B, inspector RegExp, cap config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		inspector.Inspect(ctx, cap)
	}
}

func BenchmarkRegExpByte(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range regexpTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkRegExpByte(b, test.inspector, cap)
			},
		)
	}
}
