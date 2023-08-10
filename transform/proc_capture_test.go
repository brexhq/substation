package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procCapture{}

var procCaptureTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON find",
		config.Config{
			Settings: map[string]interface{}{
				"key":        "foo",
				"set_key":    "foo",
				"type":       "find",
				"expression": "^([^@]*)@.*$",
			},
		},
		[]byte(`{"foo":"bar@qux.corge"}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"JSON find_all",
		config.Config{
			Settings: map[string]interface{}{
				"key":        "foo",
				"set_key":    "foo",
				"type":       "find_all",
				"count":      3,
				"expression": "(.{1})",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`{"foo":["b","a","r"]}`),
		},
		nil,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"type":       "find",
				"expression": "^([^@]*)@.*$",
			},
		},
		[]byte(`bar@qux.corge`),
		[][]byte{
			[]byte(`bar`),
		},
		nil,
	},
	{
		"named_group",
		config.Config{
			Settings: map[string]interface{}{
				"type":       "named_group",
				"expression": "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
			},
		},
		[]byte(`bar quux`),
		[][]byte{
			[]byte(`{"foo":"bar","qux":"quux"}`),
		},
		nil,
	},
	{
		"named_group",
		config.Config{
			Settings: map[string]interface{}{
				"key":        "capture",
				"set_key":    "capture",
				"type":       "named_group",
				"expression": "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
			},
		},
		[]byte(`{"capture":"bar quux"}`),
		[][]byte{
			[]byte(`{"capture":{"foo":"bar","qux":"quux"}}`),
		},
		nil,
	},
}

func TestProcCapture(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procCaptureTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcCapture(ctx, test.cfg)
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

func benchmarkProcCapture(b *testing.B, tform *procCapture, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcCapture(b *testing.B) {
	for _, test := range procCaptureTests {
		proc, err := newProcCapture(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcCapture(b, proc, test.test)
			},
		)
	}
}
