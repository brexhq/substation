package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &networkDomainRegisteredDomain{}

var networkDomainRegisteredDomainTests = []struct {
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
			[]byte(`b.com`),
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
			[]byte(`{"a":"b.com"}`),
		},
	},
}

func TestNetworkDomainRegisteredDomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range networkDomainRegisteredDomainTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNetworkDomainRegisteredDomain(ctx, test.cfg)
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

func benchmarkNetworkDomainRegisteredDomain(b *testing.B, tf *networkDomainRegisteredDomain, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNetworkDomainRegisteredDomain(b *testing.B) {
	for _, test := range networkDomainRegisteredDomainTests {
		tf, err := newNetworkDomainRegisteredDomain(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNetworkDomainRegisteredDomain(b, tf, test.test)
			},
		)
	}
}
