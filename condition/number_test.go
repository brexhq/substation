package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspNumber{}

var numberTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass equals",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 42,
				},
			},
		},
		[]byte(`{"foo":"42"}`),
		true,
	},
	{
		"!fail equals",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 42,
				},
			},
		},
		[]byte(`{"foo":"42"}`),
		false,
	},
	{
		"pass greater_than",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": -1,
				},
			},
		},
		[]byte(`{"foo":"0"}`),
		true,
	},
	{
		"!pass greater_than",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key":    "foo",
				"negate": true,
				"options": map[string]interface{}{
					"type":  "greater_than",
					"value": 1,
				},
			},
		},
		[]byte(`{"foo":"0"}`),
		true,
	},
	{
		"pass less_than",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "less_than",
					"value": 50,
				},
			},
		},
		[]byte(`{"foo":42}`),
		true,
	},
	{
		"pass bitwise_and",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":  "bitwise_and",
					"value": 0x0001,
				},
			},
		},
		[]byte(`{"foo":"570506001"}`),
		true,
	},
	{
		"pass data",
		config.Config{
			Type: "number",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":  "equals",
					"value": 1,
				},
			},
		},
		[]byte(`0001`),
		true,
	},
}

func TestNumber(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range numberTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspNumber(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkNumberByte(b *testing.B, inspector inspNumber, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkNumberByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range numberTests {
		insp, err := newInspNumber(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkNumberByte(b, insp, capsule)
			},
		)
	}
}
