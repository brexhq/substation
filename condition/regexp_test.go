package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspRegExp{}

var regExpTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Type: "regexp",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"expression": "^Test",
				},
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "regexp",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"expression": "^Test",
				},
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"!fail",
		config.Config{
			Type: "regexp",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"expression": "^Test",
				},
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		config.Config{
			Type: "regexp",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"expression": "ABC",
				},
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
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspRegExp(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
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
		insp, err := newInspRegExp(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkRegExpByte(b, insp, capsule)
			},
		)
	}
}
