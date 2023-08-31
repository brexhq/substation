package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modDomain{}

var modDomainTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data top_level_domain",
		config.Config{
			Settings: map[string]interface{}{
				"type": "top_level_domain",
			},
		},
		[]byte(`b.com`),
		[][]byte{
			[]byte(`com`),
		},
		nil,
	},
	{
		"data registered_domain",
		config.Config{
			Settings: map[string]interface{}{
				"type": "registered_domain",
			},
		},
		[]byte(`c.b.com`),
		[][]byte{
			[]byte(`b.com`),
		},
		nil,
	},
	{
		"data subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"type": "subdomain",
			},
		},
		[]byte(`c.b.com`),
		[][]byte{
			[]byte(`c`),
		},
		nil,
	},
	{
		"object top_level_domain",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "top_level_domain",
			},
		},
		[]byte(`{"a":"b.com"}`),
		[][]byte{
			[]byte(`{"a":"com"}`),
		},
		nil,
	},
	{
		"object registered_domain",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "registered_domain",
			},
		},
		[]byte(`{"a":"c.b.com"}`),
		[][]byte{
			[]byte(`{"a":"b.com"}`),
		},
		nil,
	},
	{
		"object subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "subdomain",
			},
		},
		[]byte(`{"a":"c.b.com"}`),
		[][]byte{
			[]byte(`{"a":"c"}`),
		},
		nil,
	},
	// empty subdomain
	{
		"no subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type":             "subdomain",
				"error_on_failure": false,
			},
		},
		[]byte(`{"a":"b.com"}`),
		[][]byte{
			[]byte(`{"a":"b.com"}`),
		},
		nil,
	},
}

func TestModDomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modDomainTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModDomain(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
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

func benchmarkModDomain(b *testing.B, tf *modDomain, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModDomain(b *testing.B) {
	for _, test := range modDomainTests {
		tf, err := newModDomain(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModDomain(b, tf, test.test)
			},
		)
	}
}
