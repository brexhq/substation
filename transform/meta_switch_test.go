package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &metaSwitch{}

var metaSwitchTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected [][]byte
}{
	{
		// This test simulates an if block by having the condition always
		// succeed.
		"if",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"object": map[string]interface{}{
											"src_key": "a",
										},
										"value": "b",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "object_copy",
							Settings: map[string]interface{}{
								"object": map[string]interface{}{
									"src_key": "a",
									"dst_key": "c",
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
		// This test simulates an if/else block by having the first condition
		// always fail and the second condition always succeed by not having
		// any conditions (the condition package will always return true if
		// there are no conditions).
		"if_else",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"object": map[string]interface{}{
											"src_key": "a",
										},
										"value": "c",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "object_copy",
							Settings: map[string]interface{}{
								"object": map[string]interface{}{
									"src_key": "a",
									"dst_key": "c",
								},
							},
						},
					},
					{
						Condition: condition.Config{},
						Transform: config.Config{
							Type: "object_copy",
							Settings: map[string]interface{}{
								"object": map[string]interface{}{
									"src_key": "a",
									"dst_key": "x",
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"b","x":"b"}`),
		},
	},
	{
		// This test simulates an if/else if block by having all conditions
		// fail. The data should be unchanged.
		"if_else_if",
		config.Config{
			Settings: map[string]interface{}{
				"cases": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"object": map[string]interface{}{
											"src_key": "a",
										},
										"value": "c",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "object_copy",
							Settings: map[string]interface{}{
								"object": map[string]interface{}{
									"src_key": "a",
									"dst_key": "c",
								},
							},
						},
					},
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "string_contains",
									Settings: map[string]interface{}{
										"object": map[string]interface{}{
											"src_key": "a",
										},
										"value": "d",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "object_copy",
							Settings: map[string]interface{}{
								"src_key": "a",
								"dst_key": "d",
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
