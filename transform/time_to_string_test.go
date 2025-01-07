package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &timeToString{}

var timeToStringTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"format": timeDefaultFmt,
			},
		},
		[]byte(`1639877490000000000`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.000Z`),
		},
	},
	{
		"data with_location",
		config.Config{
			Settings: map[string]interface{}{
				"format": timeDefaultFmt,
				// Offset from UTC by -5 hours.
				"location": "America/New_York",
			},
		},
		[]byte(`1639895490000000000`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.000Z`),
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
				"format": timeDefaultFmt,
			},
		},
		[]byte(`{"a":1639877490000000000}`),
		[][]byte{
			[]byte(`{"a":"2021-12-19T01:31:30.000Z"}`),
		},
	},
}

func TestTimeToString(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ctx := context.TODO()
	for _, test := range timeToStringTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newTimeToString(ctx, test.cfg)
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

func benchmarkTimeToString(b *testing.B, tf *timeToString, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkTimeToString(b *testing.B) {
	for _, test := range timeToStringTests {
		tf, err := newTimeToString(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTimeToString(b, tf, test.test)
			},
		)
	}
}

func FuzzTestTimeToString(f *testing.F) {
	testcases := [][]byte{
		[]byte(`1639895490`),
		[]byte(`{"a":1639877490}`),
		[]byte(`"1639895490"`),
		[]byte(`{"a":"1639877490"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newTimeToString(ctx, config.Config{
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
