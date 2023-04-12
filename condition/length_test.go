package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspLength{}

var lengthTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 3,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 3,
				},
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 4,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 4,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 4,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 4,
				},
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 3,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 3,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 2,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 2,
				},
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 3,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 3,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 3,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 3,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 4,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 4,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 2,
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 2,
				},
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"rune pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"measurement": "rune",
					"type":        "equals",
					"value":       3,
				},
			},
		},
		// 3 runes (characters), 4 bytes
		[]byte("a£c"),
		true,
	},
	{
		"array pass",
		config.Config{
			Type: "length",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"measurement": "rune",
					"type":        "equals",
					"value":       3,
				},
			},
		},
		[]byte(`{"foo":["bar",2,{"baz":"qux"}]}`),
		true,
	},
}

func TestLength(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range lengthTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspLength(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
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

func benchmarkLengthByte(b *testing.B, inspector inspLength, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkLengthByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range lengthTests {
		insp, err := newInspLength(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkLengthByte(b, insp, capsule)
			},
		)
	}
}
