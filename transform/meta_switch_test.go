package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &metaSwitch{}

var metaSwitchTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected [][]byte
}{
	// This test simulates an if block by having the condition always
	// succeed.
	{
		"if",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []map[string]interface{}{
					{
						"condition": map[string]interface{}{
							"type": "any",
							"settings": map[string]interface{}{
								"conditions": []map[string]interface{}{
									{
										"type": "string_contains",
										"settings": map[string]interface{}{
											"object": map[string]interface{}{
												"source_key": "a",
											},
											"value": "b",
										},
									},
								},
							},
						},
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "c",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","c":"b"}`),
		},
	},
	{
		"if",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []map[string]interface{}{
					{
						"condition": map[string]interface{}{
							"type": "any",
							"settings": map[string]interface{}{
								"conditions": []map[string]interface{}{
									{
										"type": "string_contains",
										"settings": map[string]interface{}{
											"object": map[string]interface{}{
												"source_key": "a",
											},
											"value": "b",
										},
									},
								},
							},
						},
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "c",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","c":"b"}`),
		},
	},
	// This test simulates an if/else block by having the first condition
	// always fail and the second condition always succeed by not having
	// any conditions (the condition package will always return true if
	// there are no conditions).
	{
		"if_else",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []map[string]interface{}{
					{
						"condition": map[string]interface{}{
							"type": "string_contains",
							"settings": map[string]interface{}{
								"object": map[string]interface{}{
									"source_key": "a",
								},
								"value": "c",
							},
						},
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "c",
									},
								},
							},
						},
					},
					{
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "z",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","z":"b"}`),
		},
	},
	{
		"if_else_if",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []map[string]interface{}{
					{
						"condition": map[string]interface{}{
							"type": "any",
							"settings": map[string]interface{}{
								"conditions": []map[string]interface{}{
									{
										"type": "string_contains",
										"settings": map[string]interface{}{
											"object": map[string]interface{}{
												"source_key": "a",
											},
											"value": "c",
										},
									},
								},
							},
						},
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "c",
									},
								},
							},
						},
					},
					{
						"condition": map[string]interface{}{
							"type": "any",
							"settings": map[string]interface{}{
								"conditions": []map[string]interface{}{
									{
										"type": "string_contains",
										"settings": map[string]interface{}{
											"object": map[string]interface{}{
												"source_key": "a",
											},
											"value": "d",
										},
									},
								},
							},
						},
						"transforms": []map[string]interface{}{
							{
								"type": "object_copy",
								"settings": map[string]interface{}{
									"object": map[string]interface{}{
										"source_key": "a",
										"target_key": "d",
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestMetaSwitch(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaSwitchTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newMetaSwitch(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.data)
			result, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Error(err)
			}

			var data [][]byte
			for _, c := range result {
				data = append(data, c.Data())
			}

			if !reflect.DeepEqual(data, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, data)
			}
		})
	}
}

func benchmarkMetaSwitch(b *testing.B, tf *metaSwitch, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkMetaSwitch(b *testing.B) {
	for _, test := range metaSwitchTests {
		tf, err := newMetaSwitch(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaSwitch(b, tf, test.data)
			},
		)
	}
}
