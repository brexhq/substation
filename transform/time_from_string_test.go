package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &timeFromString{}

var timeFromStringTests = []struct {
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
		[]byte(`2021-12-19T01:31:30.000Z`),
		[][]byte{
			[]byte(`1639877490000000000`),
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
		[]byte(`2021-12-19T01:31:30.000Z`),
		[][]byte{
			[]byte(`1639895490000000000`),
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
		[]byte(`{"a":"2021-12-19T01:31:30.000Z"}`),
		[][]byte{
			[]byte(`{"a":1639877490000000000}`),
		},
	},
}

func TestTimeFromString(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeFromStringTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newTimeFromString(ctx, test.cfg)
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

func benchmarkTimeFromString(b *testing.B, tf *timeFromString, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkTimeFromString(b *testing.B) {
	for _, test := range timeFromStringTests {
		tf, err := newTimeFromString(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTimeFromString(b, tf, test.test)
			},
		)
	}
}

func FuzzTestTimeFromString(f *testing.F) {
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
		tf, err := newTimeFromString(ctx, config.Config{
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
