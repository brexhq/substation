package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &numberGreaterThan{}

var numberGreaterThanTests = []struct {
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
				"value": 1,
			},
		},
		[]byte(`{"foo":10}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1,
			},
		},
		[]byte(`10`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 10,
			},
		},
		[]byte(`{"foo":1}`),
		false,
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
	// Floats
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1,
			},
		},
		[]byte(`1.5`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1.1,
			},
		},
		[]byte(`1.5`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": 1.5,
			},
		},
		[]byte(`{"foo":1.1}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": 1.5,
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
					"target_key": "bar",
				},
			},
		},
		[]byte(`{"foo": 100, "bar": 10}`),
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
		[]byte(`{"foo": 100, "bar": 2000}`),
		false,
	},
}

func TestNumberGreaterThan(t *testing.T) {
	ctx := context.TODO()

	for _, test := range numberGreaterThanTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newNumberGreaterThan(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Condition(ctx, message)
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

func benchmarkNumberGreaterThan(b *testing.B, insp *numberGreaterThan, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkNumberGreaterThan(b *testing.B) {
	for _, test := range numberGreaterThanTests {
		insp, err := newNumberGreaterThan(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkNumberGreaterThan(b, insp, message)
			},
		)
	}
}

func FuzzTestNumberGreaterThan(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"foo":1}`),
		[]byte(`1`),
		[]byte(`1.5`),
		[]byte(`{"foo":1.1}`),
		[]byte(`10`),
		[]byte(`-5`),
		[]byte(`0`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newNumberGreaterThan(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": 1,
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
