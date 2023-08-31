package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modCase{}

var modCaseTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data downcase",
		config.Config{
			Settings: map[string]interface{}{
				"type": "downcase",
			},
		},
		[]byte(`B`),
		[][]byte{
			[]byte(`b`),
		},
		nil,
	},
	{
		"data upcase",
		config.Config{
			Settings: map[string]interface{}{
				"type": "upcase",
			},
		},
		[]byte(`b`),
		[][]byte{
			[]byte(`B`),
		},
		nil,
	},
	{
		"data snakecase",
		config.Config{
			Settings: map[string]interface{}{
				"type": "snakecase",
			},
		},
		[]byte(`bC`),
		[][]byte{
			[]byte(`b_c`),
		},
		nil,
	},
	{
		"object downcase",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "downcase",
			},
		},
		[]byte(`{"a":"B"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
	{
		"object upcase",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "upcase",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"B"}`),
		},
		nil,
	},
	{
		"object snakecase",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "snakecase",
			},
		},
		[]byte(`{"a":"bC"})`),
		[][]byte{
			[]byte(`{"a":"b_c"})`),
		},
		nil,
	},
}

func TestModCase(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modCaseTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModCase(ctx, test.cfg)
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

func benchmarkModCase(b *testing.B, tf *modCase, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModCase(b *testing.B) {
	for _, test := range modCaseTests {
		tf, err := newModCase(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModCase(b, tf, test.test)
			},
		)
	}
}
