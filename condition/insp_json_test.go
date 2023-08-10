package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ inspector = &inspJSONValid{}

var jsonValidTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"valid",
		config.Config{},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"invalid",
		config.Config{},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"!invalid",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
			},
		},
		[]byte(`{"hello":"world"}`),
		false,
	},
	{
		"!valid",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
			},
		},
		[]byte(`{hello:"world"}`),
		true,
	},
}

func TestJSONValid(t *testing.T) {
	ctx := context.TODO()

	for _, test := range jsonValidTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.test),
			)

			insp, err := newInspJSONValid(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
	}
}

func benchmarkJSONValidByte(b *testing.B, inspector *inspJSONValid, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkJSONValidByte(b *testing.B) {
	for _, test := range jsonValidTests {
		insp, err := newInspJSONValid(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.test),
				)
				benchmarkJSONValidByte(b, insp, message)
			},
		)
	}
}
