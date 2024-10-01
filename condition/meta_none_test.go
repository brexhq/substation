package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &metaNone{}

var metaNoneTests = []struct {
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
		[]byte(`{"z":"bcd"}`),
		true,
	},
	// In this test the data is interpreted as a JSON array, as specified
	// by the source_key. This test passes because no elements in the array
	// contain "d".
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
							"value": "a",
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`{"z":["b","c","d"]}`),
		true,
	},
	// This test passes because both inspectors do not match the input.
	{
		"object_mixed",
		config.Config{
			Settings: map[string]interface{}{
				"conditions": []config.Config{
					// This inspector fails because no elements in the array contain "d".
					{
						Type: "none",
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
					// This inspector fails because the data does not match the pattern "^\\[.*\\]$".
					{
						Type: "string_match",
						Settings: map[string]interface{}{
							"pattern": "^\\[.*\\]$",
						},
					},
				},
			},
		},
		[]byte(`{"z":["a","b","c"]}`),
		true,
	},
}

func TestNoneCondition(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaNoneTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaNone(ctx, test.cfg)
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

func FuzzTestMetaNone(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"z":"a"}`),
		[]byte(`["b","b","b"]`),
		[]byte(`["a","b","c"]`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newMetaNone(ctx, config.Config{
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
