package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var logicLenGreaterThanTests = []struct {
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
					"key": "foo",
				},
				"length": 2,
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"length": 2,
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "foo",
				},
				"length": 3,
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"length": 3,
			},
		},
		[]byte(`bar`),
		false,
	},
}

func TestLogicLenGreaterThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range logicLenGreaterThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newLogicLenGreaterThan(ctx, test.cfg)
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

func benchmarkLogicLenGreaterThan(b *testing.B, insp *logicLenGreaterThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkLogicLenGreaterThan(b *testing.B) {
	for _, test := range logicLenGreaterThanTests {
		insp, err := newLogicLenGreaterThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkLogicLenGreaterThan(b, insp, message)
			},
		)
	}
}
