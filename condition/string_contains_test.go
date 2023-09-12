package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var stringContainsTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": "cd",
			},
		},
		[]byte("abcd"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"string": "CD",
			},
		},
		[]byte("abcd"),
		false,
	},
}

func TestStringContains(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringContainsTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringContains(ctx, test.cfg)
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

func benchmarkStringContains(b *testing.B, insp *stringContains, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringContains(b *testing.B) {
	for _, test := range stringContainsTests {
		insp, err := newStringContains(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringContains(b, insp, message)
			},
		)
	}
}
