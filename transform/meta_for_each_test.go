package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var metaForEachTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"meta_pipeline",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "b",
				},
				"transform": config.Config{
					Type: "meta_pipeline",
					Settings: map[string]interface{}{
						"transforms": []config.Config{
							{
								Type: "format_from_base64",
							},
							{
								Type: "compress_from_gzip",
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="]}`),
		[][]byte{
			[]byte(`{"a":["H4sIAMpcy2IA/wXAIQ0AAACAsLbY93csBiFlc4wDAAAA","H4sIAI/bzmIA/wXAMQ0AAADCMK1MAv6Pph2qjP92AwAAAA=="],"b":["foo","bar"]}`),
		},
	},
	{
		"format_from_base64",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "secrets",
					"set_key": "decoded",
				},
				"transform": config.Config{
					Type: "format_from_base64",
				},
			},
		},
		[]byte(`{"secrets":["ZHJpbms=","eW91cg==","b3ZhbHRpbmU="]}`),
		[][]byte{
			[]byte(`{"secrets":["ZHJpbms=","eW91cg==","b3ZhbHRpbmU="],"decoded":["drink","your","ovaltine"]}`),
		},
	},
	{
		"string_pattern_find",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "user_email",
					"set_key": "user_name",
				},
				"transform": config.Config{
					Type: "string_pattern_find",
					Settings: map[string]interface{}{
						"expression": "^([^@]*)@.*$",
					},
				},
			},
		},
		[]byte(`{"user_email":["john.d@example.com","jane.d@example.com"]}`),
		[][]byte{
			[]byte(`{"user_email":["john.d@example.com","jane.d@example.com"],"user_name":["john.d","jane.d"]}`),
		},
	},
	{
		"string_to_lower",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "upcase",
					"set_key": "downcase",
				},
				"transform": config.Config{
					Type: "string_to_lower",
				},
			},
		},
		[]byte(`{"upcase":["ABC","DEF"]}`),
		[][]byte{
			[]byte(`{"upcase":["ABC","DEF"],"downcase":["abc","def"]}`),
		},
	},
	{
		"network_fqdn_subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "domain",
					"set_key": "subdomain",
				},
				"transform": config.Config{
					Type: "network_fqdn_subdomain",
				},
			},
		},
		[]byte(`{"domain":["www.example.com","mail.example.top"]}`),
		[][]byte{
			[]byte(`{"domain":["www.example.com","mail.example.top"],"subdomain":["www","mail"]}`),
		},
	},
	{
		"hash_sha256",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "b",
				},
				"transform": config.Config{
					Type: "hash_sha256",
				},
			},
		},
		[]byte(`{"a":["foo","bar","baz"]}`),
		[][]byte{
			[]byte(`{"a":["foo","bar","baz"],"b":["2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae","fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9","baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096"]}`),
		},
	},
	{
		"object_insert",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "b",
				},
				"transform": config.Config{
					Type: "object_insert",
					Settings: map[string]interface{}{
						"object": map[string]interface{}{
							"set_key": "baz",
						},
						"value": "qux",
					},
				},
			},
		},
		[]byte(`{"a":[{"foo":"bar"},{"baz":"quux"}]}`),
		[][]byte{
			[]byte(`{"a":[{"foo":"bar"},{"baz":"quux"}],"b":[{"baz":"qux","foo":"bar"},{"baz":"qux"}]}`),
		},
	},
	{
		"string_replace",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "b",
				},
				"transform": config.Config{
					Type: "string_replace",
					Settings: map[string]interface{}{
						"old": "r",
						"new": "z",
					},
				},
			},
		},
		[]byte(`{"a":["bar","bard"]}`),
		[][]byte{
			[]byte(`{"a":["bar","bard"],"b":["baz","bazd"]}`),
		},
	},
	{
		"time_from_str",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "b",
				},
				"transform": config.Config{
					Type: "time_from_str",
					Settings: map[string]interface{}{
						"format": "2006-01-02T15:04:05Z",
					},
				},
			},
		},
		[]byte(`{"a":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"]}`),
		[][]byte{
			[]byte(`{"a":["2021-03-06T00:02:57Z","2021-03-06T00:03:57Z","2021-03-06T00:04:57Z"],"b":["1614988977000","1614989037000","1614989097000"]}`),
		},
	},
}

func TestForEach(t *testing.T) {
	ctx := context.TODO()
	for _, test := range metaForEachTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newMetaForEach(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
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

func benchmarkMetaForEach(b *testing.B, tf *metaForEach, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkMetaForEach(b *testing.B) {
	for _, test := range metaForEachTests {
		tf, err := newMetaForEach(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkMetaForEach(b, tf, test.test)
			},
		)
	}
}
