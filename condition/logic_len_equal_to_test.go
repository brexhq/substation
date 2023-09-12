package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var logicLenEqualToTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "a",
				},
				"length": 3,
			},
		},
		[]byte(`{"a":"bcd"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"length": 3,
			},
		},
		[]byte(`bcd`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "a",
				},
				"length": 4,
			},
		},
		[]byte(`{"a":"bcd"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"length": 4,
			},
		},
		[]byte(`bcd`),
		false,
	},
}

func TestLogicLenEqualTo(t *testing.T) {
	ctx := context.TODO()

	for _, test := range logicLenEqualToTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newLogicLenEqualTo(ctx, test.cfg)
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

func benchmarkLogicLenEqualTo(b *testing.B, insp *logicLenEqualTo, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkLogicLenEqualTo(b *testing.B) {
	for _, test := range logicLenEqualToTests {
		insp, err := newLogicLenEqualTo(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkLogicLenEqualTo(b, insp, message)
			},
		)
	}
}
