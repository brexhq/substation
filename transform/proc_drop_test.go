package transform

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procDrop{}

var procDropTests = []struct {
	name string
	cfg  config.Config
	test [][]byte
	err  error
}{
	{
		"drop",
		config.Config{},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"qux"}`),
		},
		nil,
	},
}

func TestProcDrop(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procDropTests {
		var messages []*mess.Message
		for _, data := range test.test {
			message, err := mess.New(
				mess.SetData(data),
			)
			if err != nil {
				t.Fatal(err)
			}

			messages = append(messages, message)
		}

		proc, err := newProcDrop(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}

		result, err := Apply(ctx, []Transformer{proc}, messages...)
		if err != nil {
			t.Error(err)
		}

		length := len(result)
		if length != 0 {
			t.Errorf("got %d", length)
		}
	}
}

func benchmarkProcDrop(b *testing.B, tf *procDrop, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcDrop(b *testing.B) {
	for _, test := range procDropTests {
		proc, err := newProcDrop(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcDrop(b, proc, test.test[0])
			},
		)
	}
}
