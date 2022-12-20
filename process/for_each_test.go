package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var forEachTests = []struct {
	name     string
	proc     _forEach
	test     []byte
	expected []byte
	err      error
}{
	{
		"base64",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"capture",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "capture",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"expression": "^([^@]*)@.*$",
							"type":       "find",
						},
					},
				},
			},
		},
		[]byte(`{"input":["foo@qux.com","bar@qux.com"]}`),
		[]byte(`{"input":["foo@qux.com","bar@qux.com"],"output":["foo","bar"]}`),
		nil,
	},
	{
		"case",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "case",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"type": "lowercase",
						},
					},
				},
			},
		},
		[]byte(`{"input":["ABC","DEF"]}`),
		[]byte(`{"input":["ABC","DEF"],"output":["abc","def"]}`),
		nil,
	},
	{
		"convert",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"domain",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "domain",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"type": "subdomain",
						},
					},
				},
			},
		},
		[]byte(`{"input":["www.example.com","mail.example.top"]}`),
		[]byte(`{"input":["www.example.com","mail.example.top"],"output":["www","mail"]}`),
		nil,
	},
	{
		"flatten",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "flatten",
					Settings: map[string]interface{}{
						"key":     "flatten",
						"set_key": "flatten",
						"options": map[string]interface{}{
							"deep": true,
						},
					},
				},
			},
		},
		[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}]}`),
		[]byte(`{"input":[{"flatten":[["foo"],[[["bar",[["baz"]]]]]]},{"flatten":[["foo"],[[["bar",[["baz"]]]]]]}],"output":[{"flatten":["foo","bar","baz"]},{"flatten":["foo","bar","baz"]}]}`),
		nil,
	},
	{
		"group",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "group",
					Settings: map[string]interface{}{
						"key":     "group",
						"set_key": "group",
					},
				},
			},
		},
		[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}]}`),
		[]byte(`{"input":[{"group":[["foo","bar"],[123,456]]},{"group":[["foo","bar"],[123,456]]}],"output":[{"group":[["foo",123],["bar",456]]},{"group":[["foo",123],["bar",456]]}]}`),
		nil,
	},
	{
		"hash",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"insert",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "insert",
					Settings: map[string]interface{}{
						"set_key": "baz",
						"options": map[string]interface{}{
							"value": "qux",
						},
					},
				},
			},
		},
		[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}]}`),
		[]byte(`{"input":[{"foo":"bar"},{"baz":"quux"}],"output":[{"foo":"bar","baz":"qux"},{"baz":"qux"}]}`),
		nil,
	},
	{
		"join",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "join",
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
		nil,
	},
	{
		"math",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"pipeline",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"replace",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
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
		nil,
	},
	{
		"time",
		_forEach{
			process: process{
				Key:    "input",
				SetKey: "output.-1",
			},
			Options: _forEachOptions{
				Processor: config.Config{
					Type: "time",
					Settings: map[string]interface{}{
						"options": map[string]interface{}{
							"format":     "2006-01-02T15:04:05Z",
							"set_format": "2006-01-02T15:04:05.000000Z",
						},
					},
				},
			},
		},
		[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"]}`),
		[]byte(`{"input":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"],"output":["2021-03-06T00:02:57.000000Z","2021-03-06T00:03:57.000000Z","2021-03-06T00:04:57.000000Z"]}`),
		nil,
	},
}

func TestForEach(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range forEachTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkForEach(b *testing.B, applicator _forEach, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkForEach(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range forEachTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkForEach(b, test.proc, capsule)
			},
		)
	}
}
