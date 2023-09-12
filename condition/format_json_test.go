package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var jsonValidTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{},
		[]byte(`{"hello":"world"}`),
		true,
	},
	{
		"fail",
		config.Config{},
		[]byte(`{hello:"world"}`),
		false,
	},
}

func TestFormatJSON(t *testing.T) {
	ctx := context.TODO()

	for _, test := range jsonValidTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newFormatJSON(ctx, test.cfg)
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

func benchmarkFormatJSONByte(b *testing.B, insp *formatJSON, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkFormatJSONByte(b *testing.B) {
	for _, test := range jsonValidTests {
		insp, err := newFormatJSON(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkFormatJSONByte(b, insp, message)
			},
		)
	}
}
