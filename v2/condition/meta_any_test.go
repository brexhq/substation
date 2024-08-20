package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

var _ Inspector = &metaAny{}

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
				"inspectors": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "d",
						},
					},
					{
						Type: "meta_all",
						Settings: map[string]interface{}{
							"conditions": []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"value": "b",
									},
								},
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"value": "c",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte("bcd"),
		true,
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "z",
				},
				"inspectors": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "d",
						},
					},
				},
			},
		},
		[]byte(`{"z":"bcd"}`),
		true,
	},
	{
		"object_array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "z",
				},
				"inspectors": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "d",
						},
					},
				},
			},
		},
		[]byte(`{"z":["b","c","d"]}`),
		true,
	},

	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"inspectors": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "d",
						},
					},
				},
			},
		},
		[]byte(`["b","c","d"]`),
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

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.data))
			}
		})
	}
}
