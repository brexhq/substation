package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
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
				"switch": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "insp_string",
									Settings: map[string]interface{}{
										"key":    "foo",
										"type":   "contains",
										"string": "bar",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "proc_copy",
							Settings: map[string]interface{}{
								"key":     "foo",
								"set_key": "bar",
							},
						},
					},
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`{"foo":"bar","bar":"bar"}`),
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
				"switch": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "insp_string",
									Settings: map[string]interface{}{
										"key":    "foo",
										"type":   "contains",
										"string": "bar",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "proc_copy",
							Settings: map[string]interface{}{
								"key":     "foo",
								"set_key": "bar",
							},
						},
					},
					{
						Condition: condition.Config{},
						Transform: config.Config{
							Type: "proc_copy",
							Settings: map[string]interface{}{
								"key":     "foo",
								"set_key": "baz",
							},
						},
					},
				},
			},
		},
		[]byte(`{"foo":"baz"}`),
		[][]byte{
			[]byte(`{"foo":"baz","baz":"baz"}`),
		},
	},
	{
		// This test simulates an if/else if block by having all conditions
		// fail. The data should be unchanged.
		"if_else_if",
		config.Config{
			Settings: map[string]interface{}{
				"switch": []struct {
					Condition condition.Config `json:"condition"`
					Transform config.Config    `json:"transform"`
				}{
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "insp_string",
									Settings: map[string]interface{}{
										"key":    "foo",
										"type":   "contains",
										"string": "bar",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "proc_copy",
							Settings: map[string]interface{}{
								"key":     "foo",
								"set_key": "bar",
							},
						},
					},
					{
						Condition: condition.Config{
							Operator: "any",
							Inspectors: []config.Config{
								{
									Type: "insp_string",
									Settings: map[string]interface{}{
										"key":    "foo",
										"type":   "contains",
										"string": "baz",
									},
								},
							},
						},
						Transform: config.Config{
							Type: "proc_copy",
							Settings: map[string]interface{}{
								"key":     "foo",
								"set_key": "baz",
							},
						},
					},
				},
			},
		},
		[]byte(`{"foo":"qux"}`),
		[][]byte{
			[]byte(`{"foo":"qux"}`),
		},
	},
}

func TestMetaSwitch(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaSwitchTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.data),
			)
			if err != nil {
				t.Fatal(err)
			}

			tform, err := newMetaSwitch(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := tform.Transform(ctx, message)
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

func benchmarkMetaSwitch(b *testing.B, tform *metaSwitch, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, msg)
	}
}

func BenchmarkMetaSwitch(b *testing.B) {
	for _, test := range metaSwitchTests {
		proc, err := newMetaSwitch(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaSwitch(b, proc, test.data)
			},
		)
	}
}
