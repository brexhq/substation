package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var stringLessThanTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": "b",
			},
		},
		[]byte("a"),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": "2024-01",
			},
		},
		[]byte(`2023-01-01T00:00:00Z`),
		true,
	},
}

func TestStringLessThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringLessThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringLessThan(ctx, test.cfg)
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

func benchmarkStringLessThan(b *testing.B, insp *stringLessThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringLessThan(b *testing.B) {
	for _, test := range stringLessThanTests {
		insp, err := newStringLessThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringLessThan(b, insp, message)
			},
		)
	}
}
