package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var strCaseUpTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`b`),
		[][]byte{
			[]byte(`B`),
		},
	},
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
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"B"}`),
		},
	},
}

func TestCaseUp(t *testing.T) {
	ctx := context.TODO()
	for _, test := range strCaseUpTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStrCaseUp(ctx, test.cfg)
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

func benchmarkStrCaseUp(b *testing.B, tf *strCaseUp, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStrCaseUp(b *testing.B) {
	for _, test := range strCaseUpTests {
		tf, err := newStrCaseUp(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStrCaseUp(b, tf, test.test)
			},
		)
	}
}
