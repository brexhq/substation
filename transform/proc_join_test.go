package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procJoin{}

var procJoinTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "foo",
				"set_key":   "foo",
				"separator": ".",
			},
		},
		[]byte(`{"foo":["bar","baz"]}`),
		[][]byte{
			[]byte(`{"foo":"bar.baz"}`),
		},
		nil,
	},
}

func TestProcJoin(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procJoinTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcJoin(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
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

func benchmarkProcJoin(b *testing.B, tf *procJoin, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcJoin(b *testing.B) {
	for _, test := range procJoinTests {
		proc, err := newProcJoin(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcJoin(b, proc, test.test)
			},
		)
	}
}
