package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procCopy{}
	_ Batcher  = procCopy{}
	_ Streamer = procCopy{}
)

var copyTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "baz",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":"bar","baz":"bar"}`),
		nil,
	},
	{
		"JSON unescape",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
			},
		},
		[]byte(`{"foo":"{\"bar\":\"baz\"}"`),
		[]byte(`{"foo":{"bar":"baz"}`),
		nil,
	},
	{
		"JSON unescape",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
			},
		},
		[]byte(`{"foo":"[\"bar\"]"}`),
		[]byte(`{"foo":["bar"]}`),
		nil,
	},
	{
		"from JSON",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"key": "foo",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`bar`),
		nil,
	},
	{
		"from JSON nested",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"key": "foo",
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON utf8",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"set_key": "bar",
			},
		},
		[]byte(`baz`),
		[]byte(`{"bar":"baz"}`),
		nil,
	},
	{
		"to JSON base64",
		config.Config{
			Type: "copy",
			Settings: map[string]interface{}{
				"set_key": "bar",
			},
		},
		[]byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
		[]byte(`{"bar":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
}

func TestCopy(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range copyTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcCopy(ctx, test.cfg)
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

func benchmarkCopy(b *testing.B, applier procCopy, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCopy(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range copyTests {
		proc, err := newProcCopy(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCopy(b, proc, capsule)
			},
		)
	}
}
