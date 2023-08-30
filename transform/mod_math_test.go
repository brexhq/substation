package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modMath{}

var modMathTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"add",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "math",
					"set_key": "math",
				}, "operation": "add",
			},
		},
		[]byte(`{"math":[1,3]}`),
		[][]byte{
			[]byte(`{"math":4}`),
		},
		nil,
	},
	{
		"subtract",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "math",
					"set_key": "math",
				}, "operation": "subtract",
			},
		},
		[]byte(`{"math":[5,2]}`),
		[][]byte{
			[]byte(`{"math":3}`),
		},
		nil,
	},
	{
		"multiply",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "math",
					"set_key": "math",
				}, "operation": "multiply",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[][]byte{
			[]byte(`{"math":20}`),
		},
		nil,
	},
	{
		"divide",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "math",
					"set_key": "math",
				}, "operation": "divide",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[][]byte{
			[]byte(`{"math":5}`),
		},
		nil,
	},
}

func TestModMath(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modMathTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModMath(ctx, test.cfg)
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

func benchmarkModMath(b *testing.B, tf *modMath, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModMath(b *testing.B) {
	for _, test := range modMathTests {
		tf, err := newModMath(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModMath(b, tf, test.test)
			},
		)
	}
}
