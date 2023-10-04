package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &stringEqualTo{}

var stringEqualToTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": "abcde",
			},
		},
		[]byte("abcde"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"string": "abcde",
			},
		},
		[]byte("abcdef"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"string": `""`,
			},
		},
		[]byte("\"\""),
		true,
	},
}

func TestStringEqualTo(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringEqualToTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringEqualTo(ctx, test.cfg)
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

func benchmarkStringEqualTo(b *testing.B, insp *stringEqualTo, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkStringEqualTo(b *testing.B) {
	for _, test := range stringEqualToTests {
		insp, err := newStringEqualTo(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringEqualTo(b, insp, message)
			},
		)
	}
}
