package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var logicLenLessThanTests = []struct {
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
				"length": 4,
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"length": 4,
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

func TestLogicLenLessThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range logicLenLessThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newLogicLenLessThan(ctx, test.cfg)
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

func benchmarkLogicLenLessThan(b *testing.B, insp *logicLenLessThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkLogicLenLessThan(b *testing.B) {
	for _, test := range logicLenLessThanTests {
		insp, err := newLogicLenLessThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkLogicLenLessThan(b, insp, message)
			},
		)
	}
}
