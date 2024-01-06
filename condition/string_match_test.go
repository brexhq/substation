package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &stringMatch{}

var stringMatchTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "^Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "^Test",
			},
		},
		[]byte("-Test"),
		false,
	},
}

func TestStringMatch(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringMatchTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newStringMatch(ctx, test.cfg)
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

func benchmarkStringMatchByte(b *testing.B, insp *stringMatch, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringMatchByte(b *testing.B) {
	for _, test := range stringMatchTests {
		insp, err := newStringMatch(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkStringMatchByte(b, insp, message)
			},
		)
	}
}
