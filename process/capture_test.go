package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procCapture{}
	_ Batcher  = procCapture{}
	_ Streamer = procCapture{}
)

var captureTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON find",
		config.Config{
			Type: "capture",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type":       "find",
					"expression": "^([^@]*)@.*$",
				},
			},
		},
		[]byte(`{"foo":"bar@qux.corge"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON find_all",
		config.Config{
			Type: "capture",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type":       "find_all",
					"count":      3,
					"expression": "(.{1})",
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":["b","a","r"]}`),
		nil,
	},
	{
		"data",
		config.Config{
			Type: "capture",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "find",
					"expression": "^([^@]*)@.*$",
				},
			},
		},
		[]byte(`bar@qux.corge`),
		[]byte(`bar`),
		nil,
	},
	{
		"named_group",
		config.Config{
			Type: "capture",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type":       "named_group",
					"expression": "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
				},
			},
		},
		[]byte(`bar quux`),
		[]byte(`{"foo":"bar","qux":"quux"}`),
		nil,
	},
	{
		"named_group",
		config.Config{
			Type: "capture",
			Settings: map[string]interface{}{
				"key":     "capture",
				"set_key": "capture",
				"options": map[string]interface{}{
					"type":       "named_group",
					"expression": "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
				},
			},
		},
		[]byte(`{"capture":"bar quux"}`),
		[]byte(`{"capture":{"foo":"bar","qux":"quux"}}`),
		nil,
	},
}

func TestCapture(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range captureTests {
		t.Run(test.name, func(t *testing.T) {
			proc, err := newProcCapture(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			capsule.SetData(test.test)

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

func benchmarkCapture(b *testing.B, applier procCapture, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCapture(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range captureTests {
		proc, err := newProcCapture(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCapture(b, proc, capsule)
			},
		)
	}
}
