package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

var _ Inspector = &metaNone{}

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
				"inspectors": []config.Config{
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
				"inspectors": []config.Config{
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
							"value": "a",
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
							"value": "a",
						},
					},
				},
			},
		},
		[]byte(`["b","c","d"]`),
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
