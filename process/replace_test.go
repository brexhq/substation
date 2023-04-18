package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procReplace{}
	_ Batcher = procReplace{}
)

var replaceTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"json",
		config.Config{
			Type: "replace",
			Settings: map[string]interface{}{
				"key":     "replace",
				"set_key": "replace",
				"options": map[string]interface{}{
					"old": "r",
					"new": "z",
				},
			},
		},
		[]byte(`{"replace":"bar"}`),
		[]byte(`{"replace":"baz"}`),
		nil,
	},
	{
		"json delete",
		config.Config{
			Type: "replace",
			Settings: map[string]interface{}{
				"key":     "replace",
				"set_key": "replace",
				"options": map[string]interface{}{
					"old": "z",
				},
			},
		},
		[]byte(`{"replace":"fizz"}`),
		[]byte(`{"replace":"fi"}`),
		nil,
	},
	{
		"data",
		config.Config{
			Type: "replace",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"old": "r",
					"new": "z",
				},
			},
		},
		[]byte(`bar`),
		[]byte(`baz`),
		nil,
	},
	{
		"data delete",
		config.Config{
			Type: "replace",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"old": "r",
					"new": "",
				},
			},
		},
		[]byte(`bar`),
		[]byte(`ba`),
		nil,
	},
}

func TestReplace(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range replaceTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcReplace(ctx, test.cfg)
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

func benchmarkReplace(b *testing.B, applier procReplace, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkReplace(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range replaceTests {
		proc, err := newProcReplace(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkReplace(b, proc, capsule)
			},
		)
	}
}
