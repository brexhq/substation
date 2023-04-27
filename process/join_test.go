package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier  = procJoin{}
	_ Batcher  = procJoin{}
	_ Streamer = procJoin{}
)

var joinTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		config.Config{
			Type: "join",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"separator": ".",
				},
			},
		},
		[]byte(`{"foo":["bar","baz"]}`),
		[]byte(`{"foo":"bar.baz"}`),
		nil,
	},
}

func TestJoin(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range joinTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcJoin(ctx, test.cfg)
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

func benchmarkJoin(b *testing.B, applier procJoin, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkJoin(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range joinTests {
		proc, err := newProcJoin(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkJoin(b, proc, capsule)
			},
		)
	}
}
