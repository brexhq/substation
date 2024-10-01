package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringContains{}

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
				"value": "bc",
			},
		},
		[]byte("abcd"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": "BC",
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

func benchmarkStringContains(b *testing.B, insp *stringContains, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
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

func FuzzTestStringContains(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"foo":"bar"}`),
		[]byte(`bar`),
		[]byte(`{"foo":"baz"}`),
		[]byte(`baz`),
		[]byte(`{"foo":""}`),
		[]byte(`""`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringContains(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": "bar",
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
