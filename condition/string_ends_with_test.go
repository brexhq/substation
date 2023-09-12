package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var stringEndsWithTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "a",
				},
				"string": "de",
			},
		},
		[]byte(`{"a":"bcde"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": "de",
			},
		},
		[]byte("bcde"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"string": "bc",
			},
		},
		[]byte("bcde"),
		false,
	},
}

func TestStringEndsWith(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringEndsWithTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringEndsWith(ctx, test.cfg)
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

func benchmarkStringEndsWith(b *testing.B, insp *stringEndsWith, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringEndsWith(b *testing.B) {
	for _, test := range stringEndsWithTests {
		insp, err := newStringEndsWith(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringEndsWith(b, insp, message)
			},
		)
	}
}
