package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &metaAny{}

var metaAnyTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "c",
						},
					},
				},
			},
		},
		[]byte("abc"),
		true,
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "z",
				},
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "c",
						},
					},
				},
			},
		},
		[]byte(`{"z":"abc"}`),
		true,
	},
	// In this test the data is interpreted as a JSON array, as specified
	// by the source_key. This test passes because at least one element in
	// the array contains "c".
	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "@this",
				},
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "c",
						},
					},
				},
			},
		},
		[]byte(`["a","b","c"]`),
		true,
	},
	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "@this",
				},
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "d",
						},
					},
				},
			},
		},
		[]byte(`["a","b","c"]`),
		false,
	},
	{
		"object_array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "z",
				},
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "c",
						},
					},
				},
			},
		},
		[]byte(`{"z":["a","b","c"]}`),
		true,
	},
	// This test passes because at least one inspector matches the input.
	{
		"object_mixed",
		config.Config{
			Settings: map[string]interface{}{
				"conditions": []config.Config{
					// This inspector fails because no element in the array contains "d".
					{
						Type: "any",
						Settings: map[string]interface{}{
							"object": map[string]interface{}{
								"source_key": "z",
							},
							"conditions": []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"pattern": "d",
									},
								},
							},
						},
					},
					// This inspector passes because the data matches the pattern "^{.*}$".
					{
						Type: "string_match",
						Settings: map[string]interface{}{
							"pattern": "^{.*}$",
						},
					},
				},
			},
		},
		[]byte(`{"z":["a","b","c"]}`),
		true,
	},
}

func TestAnyCondition(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaAnyTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaAny(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Condition(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.data))
			}
		})
	}
}
