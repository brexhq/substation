package transform

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modPrettyPrint{}

var modPrettyPrintTests = []struct {
	name     string
	cfg      config.Config
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"from",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[][]byte{
			[]byte(`{
				"foo":"bar"
				}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"from",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
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
		nil,
	},
	{
		"to",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "to",
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		[][]byte{
			[]byte(`{
  "foo": "bar"
}
`),
		},
		nil,
	},
}

func TestModPrettyPrint(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modPrettyPrintTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModPrettyPrint(ctx, test.cfg)
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

func benchmarkModPrettyPrint(b *testing.B, tf *modPrettyPrint, data [][]byte) {
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

func BenchmarkModPrettyPrint(b *testing.B) {
	for _, test := range modPrettyPrintTests {
		tf, err := newModPrettyPrint(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModPrettyPrint(b, tf, test.test)
			},
		)
	}
}
