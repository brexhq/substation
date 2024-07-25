package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &numberEqualTo{}

var numberEqualToTests = []struct {
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
		[]byte(`{"foo":14}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 10,
			},
		},
		[]byte(`1`),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 0,
			},
		},
		[]byte(`{"foo":0}`),
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
				"value": 1.1,
			},
		},
		[]byte(`{"foo":1.1}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1.4,
			},
		},
		[]byte(`1.4`),
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
		[]byte(`{"foo": 10, "bar": 10}`),
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
				"value": 100,
			},
		},
		[]byte(`{"foo": 100, "bar": 200}`),
		false,
	},
}

func TestNumberEqualTo(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberEqualToTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberEqualTo(ctx, test.cfg)
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

func benchmarkNumberEqualTo(b *testing.B, insp *numberEqualTo, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkNumberEqualTo(b *testing.B) {
	for _, test := range numberEqualToTests {
		insp, err := newNumberEqualTo(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberEqualTo(b, insp, message)
			},
		)
	}
}
