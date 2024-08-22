package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &numberLessThan{}

var numberLessThanTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	// Integers
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 14,
			},
		},
		[]byte(`{"foo":10}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1,
			},
		},
		[]byte(`10`),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 10,
			},
		},
		[]byte(`{"foo":1}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 5,
			},
		},
		[]byte(`15`),
		false,
	},
	// Floats
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1,
			},
		},
		[]byte(`1.5`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 0.1,
			},
		},
		[]byte(`1.5`),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 1.5,
			},
		},
		[]byte(`{"foo":1.1}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1.5,
			},
		},
		[]byte(`1`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
			},
		},
		[]byte(`{"foo": 10, "bar": 100}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
				"value": 10,
			},
		},
		[]byte(`{"foo": 2, "bar": 1}`),
		false,
	},
}

func TestNumberLessThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberLessThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberLessThan(ctx, test.cfg)
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

func benchmarkNumberLessThan(b *testing.B, insp *numberLessThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNumberLessThan(b *testing.B) {
	for _, test := range numberLessThanTests {
		insp, err := newNumberLessThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberLessThan(b, insp, message)
			},
		)
	}
}
