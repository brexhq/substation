package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &numberBitwiseAND{}

var numberBitwiseANDTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"operand": 0x0001,
			},
		},
		[]byte(`570506001`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"operand": 0x0002,
			},
		},
		[]byte(`570506001`),
		false,
	},
}

func TestNumberBitwiseAND(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberBitwiseANDTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberBitwiseAND(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
				t.Errorf("settings: %+v", test.cfg)
				t.Errorf("test: %+v", string(test.test))
			}
		})
	}
}

func benchmarkNumberBitwiseAND(b *testing.B, insp *numberBitwiseAND, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNumberBitwiseAND(b *testing.B) {
	for _, test := range numberBitwiseANDTests {
		insp, err := newNumberBitwiseAND(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberBitwiseAND(b, insp, message)
			},
		)
	}
}
