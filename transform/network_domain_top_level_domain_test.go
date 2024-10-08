package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &networkDomainTopLevelDomain{}

var networkDomainTopLevelDomainTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`b.com`),
		[][]byte{
			[]byte(`com`),
		},
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"b.com"}`),
		[][]byte{
			[]byte(`{"a":"com"}`),
		},
	},
}

func TestNetworkDomainTopLevelDomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range networkDomainTopLevelDomainTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNetworkDomainTopLevelDomain(ctx, test.cfg)
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

func benchmarkNetworkDomainTopLevelDomain(b *testing.B, tf *networkDomainTopLevelDomain, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNetworkDomainTopLevelDomain(b *testing.B) {
	for _, test := range networkDomainTopLevelDomainTests {
		tf, err := newNetworkDomainTopLevelDomain(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNetworkDomainTopLevelDomain(b, tf, test.test)
			},
		)
	}
}

func FuzzTestNetworkDomainTopLevelDomain(f *testing.F) {
	testcases := [][]byte{
		[]byte(`b.com`),
		[]byte(`{"a":"b.com"}`),
		[]byte(`example.com`),
		[]byte(`{"a":"example.com"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newNetworkDomainTopLevelDomain(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newNetworkDomainTopLevelDomain(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
