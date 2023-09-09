package transform

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var fmtFromPrettyPrintTests = []struct {
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

func TestFmtFromPrettyPrint(t *testing.T) {
	ctx := context.TODO()
	for _, test := range fmtFromPrettyPrintTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newfmtFromPrettyPrint(ctx, test.cfg)
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

func benchmarkFmtFromPrettyPrint(b *testing.B, tf *fmtFromPrettyPrint, data [][]byte) {
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

func BenchmarkFmtFromPrettyPrint(b *testing.B) {
	for _, test := range fmtFromPrettyPrintTests {
		tf, err := newfmtFromPrettyPrint(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFmtFromPrettyPrint(b, tf, test.test)
			},
		)
	}
}
