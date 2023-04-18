package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var forEachTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
	err      error
}{
	{
		"strings starts_with all",
		config.Config{
			Type: "for_each",
			Settings: map[string]interface{}{
				"key":    "input",
				"negate": false,
				"options": map[string]interface{}{
					"type": "all",
					"inspector": map[string]interface{}{
						"type": "strings",
						"settings": map[string]interface{}{
							"options": map[string]interface{}{
								"type":       "starts_with",
								"expression": "f",
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"ip private all",
		config.Config{
			Type: "for_each",
			Settings: map[string]interface{}{
				"key":    "input",
				"negate": false,
				"options": map[string]interface{}{
					"type": "all",
					"inspector": map[string]interface{}{
						"type": "ip",
						"settings": map[string]interface{}{
							"options": map[string]interface{}{
								"type": "private",
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["192.168.1.2","10.0.42.1","172.16.4.2"]}`),
		true,
		nil,
	},
	{
		"regexp any",
		config.Config{
			Type: "for_each",
			Settings: map[string]interface{}{
				"key":    "input",
				"negate": false,
				"options": map[string]interface{}{
					"type": "any",
					"inspector": map[string]interface{}{
						"type": "regexp",
						"settings": map[string]interface{}{
							"options": map[string]interface{}{
								"type": "^fizz$",
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"length none",
		config.Config{
			Type: "for_each",
			Settings: map[string]interface{}{
				"key":    "input",
				"negate": false,
				"options": map[string]interface{}{
					"type": "none",
					"inspector": map[string]interface{}{
						"type": "length",
						"settings": map[string]interface{}{
							"options": map[string]interface{}{
								"type":  "greater_than",
								"value": 7,
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"length all",
		config.Config{
			Type: "for_each",
			Settings: map[string]interface{}{
				"key":    "input",
				"negate": false,
				"options": map[string]interface{}{
					"type": "all",
					"inspector": map[string]interface{}{
						"type": "length",
						"settings": map[string]interface{}{
							"options": map[string]interface{}{
								"type":  "equals",
								"value": 4,
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
}

func TestForEach(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, tt := range forEachTests {
		t.Run(tt.name, func(t *testing.T) {
			capsule.SetData(tt.test)

			insp, err := newInspForEach(ctx, tt.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if tt.expected != check {
				t.Errorf("expected %v, got %v, %v", tt.expected, check, string(tt.test))
			}
		})
	}
}

func benchmarkForEachByte(b *testing.B, inspector inspForEach, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkForEachByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range forEachTests {
		insp, err := newInspForEach(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkForEachByte(b, insp, capsule)
			},
		)
	}
}
