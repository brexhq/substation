package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var strCaseSnakeTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`bC`),
		[][]byte{
			[]byte(`b_c`),
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
		[]byte(`{"a":"bC"})`),
		[][]byte{
			[]byte(`{"a":"b_c"})`),
		},
	},
}

func TestStrCaseSnake(t *testing.T) {
	ctx := context.TODO()
	for _, test := range strCaseSnakeTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStrCaseSnake(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Error(err)
			}

			var r [][]byte
			for _, c := range result {
				r = append(r, c.Data())
			}

			if !reflect.DeepEqual(r, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, r)
			}
		})
	}
}

func benchmarkStrCaseSnake(b *testing.B, tf *strCaseSnake, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStrCaseSnake(b *testing.B) {
	for _, test := range strCaseSnakeTests {
		tf, err := newStrCaseSnake(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStrCaseSnake(b, tf, test.test)
			},
		)
	}
}
