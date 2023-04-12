package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procGzip{}
	_ Batcher = procGzip{}
)

var gzipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"from",
		config.Config{
			Type: "gzip",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"direction": "from",
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[]byte(`foo`),
		nil,
	},
	{
		"to",
		config.Config{
			Type: "gzip",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"direction": "to",
				},
			},
		},
		[]byte(`foo`),
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		nil,
	},
}

func TestGzip(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range gzipTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcGzip(ctx, test.cfg)
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

func benchmarkGzip(b *testing.B, applier procGzip, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkGzip(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range gzipTests {
		proc, err := newProcGzip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkGzip(b, proc, capsule)
			},
		)
	}
}
