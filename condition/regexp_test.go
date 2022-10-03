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
	capsule := config.NewCapsule()

	for _, test := range regexpTests {
		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v", test.expected, check)
		}
	}
}

func benchmarkRegExpByte(b *testing.B, inspector RegExp, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkRegExpByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range regexpTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkRegExpByte(b, test.inspector, capsule)
			},
		)
	}
}
