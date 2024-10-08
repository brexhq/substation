package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &metaAll{}

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
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "a",
						},
					},
				},
			},
		},
		[]byte("a"),
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`{"z":"a"}`),
		true,
	},
	// In this test the data is interpreted as a JSON array, as specified
	// by the source_key. This test fails because not every element in the
	// array contains "a".
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`["a","a","b"]`),
		false,
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`["a","a","a"]`),
		true,
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`{"z":["a","a","a"]}`),
		true,
	},
	// This test passes because both inspectors match the input.
	{
		"object_mixed",
		config.Config{
			Settings: map[string]interface{}{
				"conditions": []config.Config{
					// This inspector passes because the elements in the array contains "a".
					{
						Type: "all",
						Settings: map[string]interface{}{
							"object": map[string]interface{}{
								"source_key": "z",
							},
							"conditions": []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"pattern": "a",
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
		[]byte(`{"z":["a","a","a"]}`),
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

func FuzzTestMetaAll(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"z":"a"}`),
		[]byte(`["a","a","b"]`),
		[]byte(`["a","a","a"]`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newMetaAll(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "@this",
				},
				"conditions": []config.Config{
					{
						Type: "string_contains",
						Settings: map[string]interface{}{
							"value": "a",
						},
					},
				},
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
