package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var regExpTests = []struct {
	name      string
	inspector inspRegExp
	test      []byte
	expected  bool
}{
	{
		"pass",
		inspRegExp{
			Options: inspRegExpOptions{
				Expression: "^Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		inspRegExp{
			Options: inspRegExpOptions{
				Expression: "^Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"!fail",
		inspRegExp{
			condition: condition{
				Negate: true,
			},
			Options: inspRegExpOptions{
				Expression: "^Test",
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		inspRegExp{
			condition: condition{
				Negate: true,
			},
			Options: inspRegExpOptions{
				Expression: "ABC",
			},
		},
		[]byte("ABC"),
		false,
	},
}

func TestRegExp(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range regExpTests {
		var _ Inspector = test.inspector

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

func benchmarkRegExpByte(b *testing.B, inspector inspRegExp, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkRegExpByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range regExpTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkRegExpByte(b, test.inspector, capsule)
			},
		)
	}
}
