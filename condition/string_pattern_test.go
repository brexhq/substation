package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var stringPatternTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"expression": "^Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"expression": "^Test",
			},
		},
		[]byte("-Test"),
		false,
	},
}

func TestStringPattern(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringPatternTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newStringPattern(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkStringPatternByte(b *testing.B, insp *stringPattern, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringPatternByte(b *testing.B) {
	for _, test := range stringPatternTests {
		insp, err := newStringPattern(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkStringPatternByte(b, insp, message)
			},
		)
	}
}
