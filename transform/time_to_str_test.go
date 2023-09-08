package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform/time"
)

var timeToStrTests = []struct {
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
				"format": time.DefaultFmt,
			},
		},
		[]byte(`1639877490000`),
		[][]byte{
			[]byte(`2021-12-19T01:31:30.000Z`),
		},
	},
	{
		"data with_location",
		config.Config{
			Settings: map[string]interface{}{
				"format": time.DefaultFmt,
				// Offset from UTC by -5 hours.
				"location": "America/New_York",
			},
		},
		[]byte(`1639895490000`),
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
					"key":     "a",
					"set_key": "a",
				},
				"format": time.DefaultFmt,
			},
		},
		[]byte(`{"a":1639877490000}`),
		[][]byte{
			[]byte(`{"a":"2021-12-19T01:31:30.000Z"}`),
		},
	},
}

func TestTimeToStr(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeToStrTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newTimeToStr(ctx, test.cfg)
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

func benchmarkTimeToStr(b *testing.B, tf *timeToStr, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkTimeToStr(b *testing.B) {
	for _, test := range timeToStrTests {
		tf, err := newTimeToStr(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTimeToStr(b, tf, test.test)
			},
		)
	}
}
