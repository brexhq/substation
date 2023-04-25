package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procConvert{}
	_ Batcher  = procConvert{}
	_ Streamer = procConvert{}
)

var convertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"bool true",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "bool",
				},
			},
		},
		[]byte(`{"foo":"true"}`),
		[]byte(`{"foo":true}`),
		nil,
	},
	{
		"bool false",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "bool",
				},
			},
		},
		[]byte(`{"foo":"false"}`),
		[]byte(`{"foo":false}`),
		nil,
	},
	{
		"int",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "int",
				},
			},
		},
		[]byte(`{"foo":"-123"}`),
		[]byte(`{"foo":-123}`),
		nil,
	},
	{
		"float",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "float",
				},
			},
		},
		[]byte(`{"foo":"123.456"}`),
		[]byte(`{"foo":123.456}`),
		nil,
	},
	{
		"uint",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "uint",
				},
			},
		},
		[]byte(`{"foo":"123"}`),
		[]byte(`{"foo":123}`),
		nil,
	},
	{
		"string",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "string",
				},
			},
		},
		[]byte(`{"foo":123}`),
		[]byte(`{"foo":"123"}`),
		nil,
	},
	{
		"int",
		config.Config{
			Type: "convert",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "int",
				},
			},
		},
		[]byte(`{"foo":123.456}`),
		[]byte(`{"foo":123}`),
		nil,
	},
}

func TestConvert(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range convertTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcConvert(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
	}
}

func benchmarkConvert(b *testing.B, applier procConvert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkConvert(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range convertTests {
		proc, err := newProcConvert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkConvert(b, proc, capsule)
			},
		)
	}
}
