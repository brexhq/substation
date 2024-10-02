package transform

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &formatFromPrettyPrint{}

var formatFromPrettyPrintTests = []struct {
	name     string
	cfg      config.Config
	test     [][]byte
	expected [][]byte
}{
	{
		"from",
		config.Config{},
		[][]byte{
			[]byte(`{
				"foo":"bar"
				}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
	},
	{
		"from",
		config.Config{},
		[][]byte{
			[]byte(`{`),
			[]byte(`"foo":"bar",`),
			[]byte(`"baz": {`),
			[]byte(`	"qux": "corge"`),
			[]byte(`}`),
			[]byte(`}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar","baz":{"qux":"corge"}}`),
		},
	},
}

func TestFormatFromPrettyPrint(t *testing.T) {
	ctx := context.TODO()
	for _, test := range formatFromPrettyPrintTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newFormatFromPrettyPrint(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			var messages []*message.Message
			for _, data := range test.test {
				msg := message.New().SetData(data)
				messages = append(messages, msg)
			}

			result, err := Apply(ctx, []Transformer{tf}, messages...)
			if err != nil {
				t.Error(err)
			}

			for i, res := range result {
				expected := test.expected[i]
				if !bytes.Equal(expected, res.Data()) {
					t.Errorf("expected %s, got %s", expected, res.Data())
				}
			}
		})
	}
}

func benchmarkFormatFromPrettyPrint(b *testing.B, tf *formatFromPrettyPrint, data [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		var messages []*message.Message
		for _, d := range data {
			msg := message.New().SetData(d)
			messages = append(messages, msg)
		}

		_, _ = Apply(ctx, []Transformer{tf}, messages...)
	}
}

func BenchmarkFormatFromPrettyPrint(b *testing.B) {
	for _, test := range formatFromPrettyPrintTests {
		tf, err := newFormatFromPrettyPrint(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFormatFromPrettyPrint(b, tf, test.test)
			},
		)
	}
}

func FuzzTestFormatFromPrettyPrint(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"c":"d"}`),
		[]byte(`{"e":"f"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		tf, err := newFormatFromPrettyPrint(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
