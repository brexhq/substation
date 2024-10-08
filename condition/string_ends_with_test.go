package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringEndsWith{}

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
					"source_key": "a",
				},
				"value": "de",
			},
		},
		[]byte(`{"a":"bcde"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "de",
			},
		},
		[]byte("bcde"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": "bc",
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

			check, err := insp.Condition(ctx, message)
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
		_, _ = insp.Condition(ctx, message)
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

func FuzzTestStringEndsWith(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"bcde"}`),
		[]byte(`bcde`),
		[]byte(`{"a":"abcd"}`),
		[]byte(`abcd`),
		[]byte(`{"a":""}`),
		[]byte(`""`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringEndsWith(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": "de",
			},
		})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
