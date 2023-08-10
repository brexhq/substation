package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ inspector = &inspString{}

var stringTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"key":    "foo",
				"type":   "starts_with",
				"string": "Test",
			},
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "starts_with",
				"string": "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "starts_with",
				"string": "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "equals",
				"string": "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "equals",
				"string": "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "contains",
				"string": "es",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "contains",
				"string": "ABC",
			},
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
				"type":   "starts_with",
				"string": "XYZ",
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
				"type":   "starts_with",
				"string": "ABC",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
				"type":   "equals",
				"string": "",
			},
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		config.Config{
			Settings: map[string]interface{}{
				"negate": true,
				"type":   "contains",
				"string": "A",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "equals",
				"string": `""`,
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"key":    "foo",
				"type":   "equals",
				"string": "",
			},
		},
		[]byte(``),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "greater_than",
				"string": "a",
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"type":   "less_than",
				"string": "c",
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"key":    "a",
				"type":   "greater_than",
				"string": "2022-01-01T00:00:00Z",
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"key":    "a",
				"type":   "less_than",
				"string": "2024-01",
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
}

func TestString(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.data),
			)

			insp, err := newInspString(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkStringByte(b *testing.B, inspector *inspString, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkStringByte(b *testing.B) {
	for _, test := range stringTests {
		insp, err := newInspString(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.data),
				)
				benchmarkStringByte(b, insp, message)
			},
		)
	}
}
