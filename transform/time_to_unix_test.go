package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &timeToUnix{}

var timeToUnixTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{},
		},
		[]byte(`1639895490000000000`),
		[][]byte{
			[]byte(`1639895490`),
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
		[]byte(`{"a":1639877490000000000}`),
		[][]byte{
			[]byte(`{"a":1639877490}`),
		},
	},
}

func TestTimeToUnix(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeToUnixTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newTimeToUnix(ctx, test.cfg)
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

func benchmarkTimeToUnix(b *testing.B, tf *timeToUnix, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkTimeToUnix(b *testing.B) {
	for _, test := range timeToUnixTests {
		tf, err := newTimeToUnix(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTimeToUnix(b, tf, test.test)
			},
		)
	}
}

func FuzzTestTimeToUnixSeconds(f *testing.F) {
	testcases := [][]byte{
		[]byte(`"2023-01-01T00:00:00Z"`),
		[]byte(`"2023-01-01 00:00:00"`),
		[]byte(`"01/01/2023"`),
		[]byte(`"2023-01-01"`),
		[]byte(`"2023-01-01T00:00:00+00:00"`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newTimeToUnix(ctx, config.Config{
			Settings: map[string]interface{}{
				"layout": "2006-01-02T15:04:05Z07:00",
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
