package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procConvert{}

var procConvertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"bool true",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "bool",
			},
		},
		[]byte(`{"foo":"true"}`),
		[][]byte{
			[]byte(`{"foo":true}`),
		},
		nil,
	},
	{
		"bool false",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "bool",
			},
		},
		[]byte(`{"foo":"false"}`),
		[][]byte{
			[]byte(`{"foo":false}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "int",
			},
		},
		[]byte(`{"foo":"-123"}`),
		[][]byte{
			[]byte(`{"foo":-123}`),
		},
		nil,
	},
	{
		"float",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "float",
			},
		},
		[]byte(`{"foo":"123.456"}`),
		[][]byte{
			[]byte(`{"foo":123.456}`),
		},
		nil,
	},
	{
		"uint",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "uint",
			},
		},
		[]byte(`{"foo":"123"}`),
		[][]byte{
			[]byte(`{"foo":123}`),
		},
		nil,
	},
	{
		"string",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "string",
			},
		},
		[]byte(`{"foo":123}`),
		[][]byte{
			[]byte(`{"foo":"123"}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Type: "proc_convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "int",
			},
		},
		[]byte(`{"foo":123.456}`),
		[][]byte{
			[]byte(`{"foo":123}`),
		},
		nil,
	},
}

func TestProcConvert(t *testing.T) {
	ctx := context.TODO()

	for _, test := range procConvertTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcConvert(ctx, test.cfg)
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

func benchmarkProcConvert(b *testing.B, tform *procConvert, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcConvert(b *testing.B) {
	for _, test := range procConvertTests {
		proc, err := newProcConvert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcConvert(b, proc, test.test)
			},
		)
	}
}
