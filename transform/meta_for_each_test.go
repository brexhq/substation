package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &metaForEach{}

var metaForEachTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"base64",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_base64",
					Settings: map[string]interface{}{
						"direction": "from",
					},
				},
			},
		},
		[]byte(`{"input":["Zm9v","YmFy"]}`),
		[][]byte{
			[]byte(`{"input":["Zm9v","YmFy"],"output":["foo","bar"]}`),
		},
		nil,
	},
	{
		"capture",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_capture",
					Settings: map[string]interface{}{
						"expression": "^([^@]*)@.*$",
						"type":       "find",
					},
				},
			},
		},
		[]byte(`{"input":["foo@qux.com","bar@qux.com"]}`),
		[][]byte{
			[]byte(`{"input":["foo@qux.com","bar@qux.com"],"output":["foo","bar"]}`),
		},
		nil,
	},
	{
		"case",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_case",
					Settings: map[string]interface{}{
						"type": "lower",
					},
				},
			},
		},
		[]byte(`{"input":["ABC","DEF"]}`),
		[][]byte{
			[]byte(`{"input":["ABC","DEF"],"output":["abc","def"]}`),
		},
		nil,
	},
	{
		"convert",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_convert",
					Settings: map[string]interface{}{
						"type": "bool",
					},
				},
			},
		},
		[]byte(`{"input":["true","false"]}`),
		[][]byte{
			[]byte(`{"input":["true","false"],"output":[true,false]}`),
		},
		nil,
	},
	{
		"domain",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_domain",
					Settings: map[string]interface{}{
						"type": "subdomain",
					},
				},
			},
		},
		[]byte(`{"input":["www.example.com","mail.example.top"]}`),
		[][]byte{
			[]byte(`{"input":["www.example.com","mail.example.top"],"output":["www","mail"]}`),
		},
		nil,
	},
	{
		"flatten",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_flatten",
					Settings: map[string]interface{}{
						"key":     "flatten",
						"set_key": "flatten",

						"deep": true,
					},
				},
			},
		},
		[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}]}`),
		[][]byte{
			[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}],"output":[{"flatten":["foo","bar","baz"]},{"flatten":["foo","bar","baz"]}]}`),
		},
		nil,
	},
	{
		"group",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_group",
					Settings: map[string]interface{}{
						"key":     "group",
						"set_key": "group",
					},
				},
			},
		},
		[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}]}`),
		[][]byte{
			[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}],"output":[{"group":[["foo",123],["bar",456]]},{"group":[["foo",123],["bar",456]]}]}`),
		},
		nil,
	},
	{
		"hash",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_hash",
					Settings: map[string]interface{}{
						"algorithm": "sha256",
					},
				},
			},
		},
		[]byte(`{"input":["foo","bar","baz"]}`),
		[][]byte{
			[]byte(`{"input":["foo","bar","baz"],"output":["2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae","fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9","baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096"]}`),
		},
		nil,
	},
	{
		"insert",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_insert",
					Settings: map[string]interface{}{
						"set_key": "baz",

						"value": "qux",
					},
				},
			},
		},
		[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}]}`),
		[][]byte{
			[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}],"output":[{"foo":"bar","baz":"qux"},{"baz":"qux"}]}`),
		},
		nil,
	},
	{
		"join",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_join",
					Settings: map[string]interface{}{
						"separator": ".",
					},
				},
			},
		},
		[]byte(`{"input":[["foo","bar"],["baz","qux"]]}`),
		[][]byte{
			[]byte(`{"input":[["foo","bar"],["baz","qux"]],"output":["foo.bar","baz.qux"]}`),
		},
		nil,
	},
	{
		"math",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_math",
					Settings: map[string]interface{}{
						"operation": "add",
					},
				},
			},
		},
		[]byte(`{"input":[[2,3],[4,5]]}`),
		[][]byte{
			[]byte(`{"input":[[2,3],[4,5]],"output":[5,9]}`),
		},
		nil,
	},
	{
		"pipeline",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "meta_pipeline",
					Settings: map[string]interface{}{
						"transforms": []config.Config{
							{
								Type: "proc_base64",
								Settings: map[string]interface{}{
									"direction": "from",
								},
							},
							{
								Type: "proc_gzip",
								Settings: map[string]interface{}{
									"direction": "from",
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="]}`),
		[][]byte{
			[]byte(`{"input":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="],"output":["foo","bar"]}`),
		},
		nil,
	},
	{
		"replace",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_replace",
					Settings: map[string]interface{}{
						"old": "r",
						"new": "z",
					},
				},
			},
		},
		[]byte(`{"input":["bar","bard"]}`),
		[][]byte{
			[]byte(`{"input":["bar","bard"],"output":["baz","bazd"]}`),
		},
		nil,
	},
	{
		"time",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "input",
				"set_key": "output.-1",
				"transform": config.Config{
					Type: "proc_time",
					Settings: map[string]interface{}{
						"format":     "2006-01-02T15:04:05Z",
						"set_format": "2006-01-02T15:04:05.000000Z",
					},
				},
			},
		},
		[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"]}`),
		[][]byte{
			[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"],"output":["2021-03-06T00:02:57.000000Z","2021-03-06T00:03:57.000000Z","2021-03-06T00:04:57.000000Z"]}`),
		},
		nil,
	},
}

func TestForEach(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaForEachTests {
		t.Run(test.name, func(t *testing.T) {
			capsule, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newMetaForEach(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, capsule)
			if err != nil {
				t.Fatal(err)
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

func benchmarkMetaForEach(b *testing.B, tform *metaForEach, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, msg)
	}
}

func BenchmarkMetaForEach(b *testing.B) {
	for _, test := range metaForEachTests {
		proc, err := newMetaForEach(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaForEach(b, proc, test.test)
			},
		)
	}
}
