package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Inspector = &inspStrings{}

var stringsTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"key":        "foo",
				"type":       "starts_with",
				"expression": "Test",
			},
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "starts_with",
				"expression": "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "starts_with",
				"expression": "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "equals",
				"expression": "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "equals",
				"expression": "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "contains",
				"expression": "es",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "contains",
				"expression": "ABC",
			},
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"negate":     true,
				"type":       "starts_with",
				"expression": "XYZ",
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"negate":     true,
				"type":       "starts_with",
				"expression": "ABC",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"negate":     true,
				"type":       "equals",
				"expression": "",
			},
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"negate":     true,
				"type":       "contains",
				"expression": "A",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "equals",
				"expression": `""`,
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"key":        "foo",
				"type":       "equals",
				"expression": "",
			},
		},
		[]byte(``),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "greater_than",
				"expression": "a",
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"type":       "less_than",
				"expression": "c",
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"key":        "a",
				"type":       "greater_than",
				"expression": "2022-01-01T00:00:00Z",
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "insp_strings",
			Settings: map[string]interface{}{
				"key":        "a",
				"type":       "less_than",
				"expression": "2024-01",
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
}

func TestStrings(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringsTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.data),
			)

			insp, err := newInspStrings(ctx, test.cfg)
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

func benchmarkStringsByte(b *testing.B, inspector *inspStrings, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkStringsByte(b *testing.B) {
	for _, test := range stringsTests {
		insp, err := newInspStrings(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.data),
				)
				benchmarkStringsByte(b, insp, message)
			},
		)
	}
}
