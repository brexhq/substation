package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/internal/config"
)

var foreachTests = []struct {
	name     string
	proc     ForEach
	test     []byte
	expected []byte
}{
	{
		"base64",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "base64",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"direction": "from",
						},
					},
				},
			},
		},
		[]byte(`{"input":["Zm9v","YmFy"]}`),
		[]byte(`{"input":["Zm9v","YmFy"],"output":["foo","bar"]}`),
	},
	{
		"capture",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "capture",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"expression": "^([^@]*)@.*$",
							"function":   "find",
						},
					},
				},
			},
		},
		[]byte(`{"input":["foo@qux.com","bar@qux.com"]}`),
		[]byte(`{"input":["foo@qux.com","bar@qux.com"],"output":["foo","bar"]}`),
	},
	{
		"case",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "case",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"case": "lower",
						},
					},
				},
			},
		},
		[]byte(`{"input":["ABC","DEF"]}`),
		[]byte(`{"input":["ABC","DEF"],"output":["abc","def"]}`),
	},
	{
		"concat",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "concat",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"separator": ".",
						},
					},
				},
			},
		},
		[]byte(`{"input":[["foo","bar"],["baz","qux"]]}`),
		[]byte(`{"input":[["foo","bar"],["baz","qux"]],"output":["foo.bar","baz.qux"]}`),
	},
	{
		"convert",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "convert",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"type": "bool",
						},
					},
				},
			},
		},
		[]byte(`{"input":["true","false"]}`),
		[]byte(`{"input":["true","false"],"output":[true,false]}`),
	},
	{
		"domain",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "domain",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"function": "subdomain",
						},
					},
				},
			},
		},
		[]byte(`{"input":["www.example.com","mail.example.top"]}`),
		[]byte(`{"input":["www.example.com","mail.example.top"],"output":["www","mail"]}`),
	},
	{
		"flatten",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "flatten",
					Settings: map[string]interface{}{
						"input_key":  "flatten",
						"output_key": "flatten",
						"options": map[string]interface{}{
							"deep": true,
						},
					},
				},
			},
		},
		[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}]}`),
		[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}],"output":[{"flatten":["foo","bar","baz"]},{"flatten":["foo","bar","baz"]}]}`),
	},
	{
		"group",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "group",
					Settings: map[string]interface{}{
						"input_key":  "group",
						"output_key": "group",
					},
				},
			},
		},
		[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}]}`),
		[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}],"output":[{"group":[["foo",123],["bar",456]]},{"group":[["foo",123],["bar",456]]}]}`),
	},
	{
		"hash",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "hash",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"algorithm": "sha256",
						},
					},
				},
			},
		},
		[]byte(`{"input":["foo","bar","baz"]}`),
		[]byte(`{"input":["foo","bar","baz"],"output":["2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae","fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9","baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096"]}`),
	},
	{
		"insert",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "insert",
					Settings: map[string]interface{}{
						"output_key": "baz",
						"options": map[string]interface{}{
							"value": "qux",
						},
					},
				},
			},
		},
		[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}]}`),
		[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}],"output":[{"foo":"bar","baz":"qux"},{"baz":"qux"}]}`),
	},
	{
		"math",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "math",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"operation": "add",
						},
					},
				},
			},
		},
		[]byte(`{"input":[[2,3],[4,5]]}`),
		[]byte(`{"input":[[2,3],[4,5]],"output":[5,9]}`),
	},
	{
		"pipeline",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "pipeline",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"processors": []config.Config{
								{
									Type: "base64",
									Settings: map[string]interface{}{
										"options": map[string]interface{}{
											"direction": "from",
										},
									},
								},
								{
									Type: "gzip",
									Settings: map[string]interface{}{
										"options": map[string]interface{}{
											"direction": "from",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		[]byte(`{"input":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="]}`),
		[]byte(`{"input":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="],"output":["foo","bar"]}`),
	},
	{
		"replace",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "replace",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"old": "r",
							"new": "z",
						},
					},
				},
			},
		},
		[]byte(`{"input":["bar","bard"]}`),
		[]byte(`{"input":["bar","bard"],"output":["baz","bazd"]}`),
	},
	{
		"time",
		ForEach{
			InputKey:  "input",
			OutputKey: "output.-1",
			Options: ForEachOptions{
				Processor: config.Config{
					Type: "time",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"input_format":  "2006-01-02T15:04:05Z",
							"output_format": "2006-01-02T15:04:05.000000Z",
						},
					},
				},
			},
		},
		[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"]}`),
		[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"],"output":["2021-03-06T00:02:57.000000Z","2021-03-06T00:03:57.000000Z","2021-03-06T00:04:57.000000Z"]}`),
	},
}

func TestForEach(t *testing.T) {
	for _, test := range foreachTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkForEachByte(b *testing.B, byter ForEach, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkForEachByte(b *testing.B) {
	for _, test := range foreachTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkForEachByte(b, test.proc, test.test)
			},
		)
		break
	}
}
