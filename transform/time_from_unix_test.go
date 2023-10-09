package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &timeFromUnix{}

var timeUnixFromTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{},
		},
		[]byte(`1639895490`),
		[][]byte{
			[]byte(`1639895490000000000`),
		},
		nil,
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
		[]byte(`{"a":1639877490}`),
		[][]byte{
			[]byte(`{"a":1639877490000000000}`),
		},
		nil,
	},
}

func TestTimeFromUnix(t *testing.T) {
	ctx := context.TODO()
	for _, test := range timeUnixFromTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newTimeFromUnix(ctx, test.cfg)
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

func benchmarkTimeFromUnix(b *testing.B, tf *timeFromUnix, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkTimeFromUnix(b *testing.B) {
	for _, test := range timeUnixFromTests {
		tf, err := newTimeFromUnix(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTimeFromUnix(b, tf, test.test)
			},
		)
	}
}
