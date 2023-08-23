package transform

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procPrettyPrint{}

var procPrettyPrintTests = []struct {
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

func TestProcPrettyPrint(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procPrettyPrintTests {
		t.Run(test.name, func(t *testing.T) {
			var messages []*mess.Message
			for _, tt := range test.test {
				message, err := mess.New(
					mess.SetData(tt),
				)
				if err != nil {
					t.Fatal(err)
				}

				messages = append(messages, message)
			}

			proc, err := newProcPrettyPrint(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := Apply(ctx, []Transformer{proc}, messages...)
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

func benchmarkProcPrettyPrint(b *testing.B, tf *procPrettyPrint, data [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		var messages []*mess.Message
		for _, d := range data {
			message, _ := mess.New(
				mess.SetData(d),
			)
			messages = append(messages, message)
		}

		_, _ = Apply(ctx, []Transformer{tf}, messages...)
	}
}

func BenchmarkProcPrettyPrint(b *testing.B) {
	for _, test := range procPrettyPrintTests {
		proc, err := newProcPrettyPrint(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcPrettyPrint(b, proc, test.test)
			},
		)
	}
}
