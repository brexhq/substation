package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procBase64{}
	_ Batcher = procBase64{}
)

var base64Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"data decode",
		config.Config{
			Type: "base64",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"direction": "from",
				},
			},
		},
		[]byte(`YmFy`),
		[]byte(`bar`),
		nil,
	},
	{
		"data encode",
		config.Config{
			Type: "base64",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"direction": "to",
				},
			},
		},
		[]byte(`bar`),
		[]byte(`YmFy`),
		nil,
	},
	{
		"JSON decode",
		config.Config{
			Type: "base64",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"direction": "from",
				},
			},
		},
		[]byte(`{"foo":"YmFy"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
}

func TestBase64(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range base64Tests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcBase64(test.cfg)
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

func benchmarkbase64(b *testing.B, applier procBase64, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkBase64(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range base64Tests {
		proc, err := newProcBase64(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkbase64(b, proc, capsule)
			},
		)
	}
}
