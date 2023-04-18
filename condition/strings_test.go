package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspStrings{}

var stringsTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":       "starts_with",
					"expression": "Test",
				},
			},
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "starts_with",
					"expression": "Test",
				},
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "starts_with",
					"expression": "Test",
				},
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "equals",
					"expression": "Test",
				},
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "equals",
					"expression": "Test",
				},
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "es",
				},
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "ABC",
				},
			},
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":       "starts_with",
					"expression": "XYZ",
				},
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":       "starts_with",
					"expression": "ABC",
				},
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":       "equals",
					"expression": "",
				},
			},
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type":       "contains",
					"expression": "A",
				},
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "equals",
					"expression": `""`,
				},
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"key": "foo",
				"options": map[string]interface{}{
					"type":       "equals",
					"expression": "",
				},
			},
		},
		[]byte(``),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "greater_than",
					"expression": "a",
				},
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "less_than",
					"expression": "c",
				},
			},
		},
		[]byte("b"),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"key": "a",
				"options": map[string]interface{}{
					"type":       "greater_than",
					"expression": "2022-01-01T00:00:00Z",
				},
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
	{
		"pass",
		config.Config{
			Type: "strings",
			Settings: map[string]interface{}{
				"key": "a",
				"options": map[string]interface{}{
					"type":       "less_than",
					"expression": "2024-01",
				},
			},
		},
		[]byte(`{"a":"2023-01-01T00:00:00Z"}`),
		true,
	},
}

func TestStrings(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range stringsTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspStrings(ctx, test.cfg)
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

func benchmarkStringsByte(b *testing.B, inspector inspStrings, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkStringsByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range stringsTests {
		insp, err := newInspStrings(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkStringsByte(b, insp, capsule)
			},
		)
	}
}
