package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &networkDomainSubdomain{}

var networkDomainSubdomainTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`c.b.com`),
		[][]byte{
			[]byte(`c`),
		},
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":"c.b.com"}`),
		[][]byte{
			[]byte(`{"a":"c"}`),
		},
	},
}

func TestNetworkDomainSubdomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range networkDomainSubdomainTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNetworkDomainSubdomain(ctx, test.cfg)
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

func benchmarkNetworkDomainSubdomain(b *testing.B, tf *networkDomainSubdomain, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNetworkDomainSubdomain(b *testing.B) {
	for _, test := range networkDomainSubdomainTests {
		tf, err := newNetworkDomainSubdomain(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNetworkDomainSubdomain(b, tf, test.test)
			},
		)
	}
}
