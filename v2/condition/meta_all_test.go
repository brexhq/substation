package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

var _ Inspector = &metaAll{}

var metaAllTests = []struct {
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
							"value": "bbb",
						},
					},
					{
						Type: "meta_any",
						Settings: map[string]interface{}{
							"inspectors": []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"value": "a",
									},
								},
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"value": "b",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte("bbb"),
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
							"value": "b",
						},
					},
				},
			},
		},
		[]byte(`{"z":"bbb"}`),
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
							"value": "b",
						},
					},
				},
			},
		},
		[]byte(`["b","b","b"]`),
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
							"value": "b",
						},
					},
				},
			},
		},
		[]byte(`{"z":["b","b","b"]}`),
		true,
	},
}

func TestAllCondition(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaAllTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaAll(ctx, test.cfg)
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
