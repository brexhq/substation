package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringStartsWith{}

var stringStartsWithTests = []struct {
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
				"value": "bc",
			},
		},
		[]byte(`{"a":"bcde"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "bc",
			},
		},
		[]byte("bcde"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": "de",
			},
		},
		[]byte("bcde"),
		false,
	},
}

func TestStringStartsWith(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringStartsWithTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringStartsWith(ctx, test.cfg)
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

func benchmarkStringStartsWith(b *testing.B, insp *stringStartsWith, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkStringStartsWith(b *testing.B) {
	for _, test := range stringStartsWithTests {
		insp, err := newStringStartsWith(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringStartsWith(b, insp, message)
			},
		)
	}
}

func FuzzTestStringStartsWith(f *testing.F) {
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
		insp, err := newStringStartsWith(ctx, config.Config{
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
