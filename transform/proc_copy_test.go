package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procCopy{}

var procCopyTests = []struct {
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
				"key":     "foo",
				"set_key": "baz",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`{"foo":"bar","baz":"bar"}`),
		},
		nil,
	},
	{
		"JSON unescape",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
			},
		},
		[]byte(`{"foo":"{\"bar\":\"baz\"}"`),
		[][]byte{
			[]byte(`{"foo":{"bar":"baz"}`),
		},
		nil,
	},
	{
		"JSON unescape",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
			},
		},
		[]byte(`{"foo":"[\"bar\"]"}`),
		[][]byte{
			[]byte(`{"foo":["bar"]}`),
		},
		nil,
	},
	{
		"from JSON",
		config.Config{
			Settings: map[string]interface{}{
				"key": "foo",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`bar`),
		},
		nil,
	},
	{
		"from JSON nested",
		config.Config{
			Settings: map[string]interface{}{
				"key": "foo",
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[][]byte{
			[]byte(`{"bar":"baz"}`),
		},
		nil,
	},
	{
		"to JSON utf8",
		config.Config{
			Settings: map[string]interface{}{
				"set_key": "bar",
			},
		},
		[]byte(`baz`),
		[][]byte{
			[]byte(`{"bar":"baz"}`),
		},
		nil,
	},
	{
		"to JSON base64",
		config.Config{
			Settings: map[string]interface{}{
				"set_key": "bar",
			},
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[][]byte{
			[]byte(`{"bar":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		},
		nil,
	},
}

func TestProcCopy(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procCopyTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcCopy(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
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

func benchmarkProcCopy(b *testing.B, tf *procCopy, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, err := mess.New(
			mess.SetData(data),
		)
		if err != nil {
			b.Fatal(err)
		}

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkProcCopy(b *testing.B) {
	for _, test := range procCopyTests {
		proc, err := newProcCopy(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcCopy(b, proc, test.test)
			},
		)
	}
}
