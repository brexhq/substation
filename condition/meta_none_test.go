package condition

import (
	"context"
	"fmt"
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
	// This should fail, but doesn't.
	{
		"object_missing",
		config.Config{
			Settings: map[string]interface{}{
				"conditions": []config.Config{
					{
						Type: "none",
						Settings: map[string]interface{}{
							"object": map[string]interface{}{
								"source_key": "z",
							},
							"conditions": []config.Config{
								{
									Type: "string_equal_to",
									Settings: map[string]interface{}{
										"object": map[string]interface{}{
											"source_key": "a",
										},
										"value": "b",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"z":[{"a":"b"}]}`),
		false,
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

			fmt.Println("got:", check) // Debugging output.

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
