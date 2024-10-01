package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &formatJSON{}

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
		"pass",
		config.Config{},
		[]byte(`["a","b","c"]`),
		true,
	},
	{
		"fail",
		config.Config{},
		[]byte(`{hello:"world"}`),
		false,
	},
	{
		"fail",
		config.Config{},
		[]byte(`a`),
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

			check, err := insp.Condition(ctx, message)
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
		_, _ = insp.Condition(ctx, message)
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

func FuzzTestFormatJSON(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"hello":"world"}`),
		[]byte(`["a","b","c"]`),
		[]byte(`{hello:"world"}`),
		[]byte(`a`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newFormatJSON(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
