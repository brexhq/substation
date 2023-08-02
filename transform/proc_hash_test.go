package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procHash{}

var procHashTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON md5",
		config.Config{
			Type: "proc_hash",
			Settings: map[string]interface{}{
				"key":       "hash",
				"set_key":   "hash",
				"algorithm": "md5",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[][]byte{
			[]byte(`{"hash":"acbd18db4cc2f85cedef654fccc4a4d8"}`),
		},
		nil,
	},
	{
		"JSON sha256",
		config.Config{
			Type: "proc_hash",
			Settings: map[string]interface{}{
				"key":       "hash",
				"set_key":   "hash",
				"algorithm": "sha256",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[][]byte{
			[]byte(`{"hash":"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"}`),
		},
		nil,
	},
	{
		"JSON @this sha256",
		config.Config{
			Type: "proc_hash",
			Settings: map[string]interface{}{
				"key":       "@this",
				"set_key":   "hash",
				"algorithm": "sha256",
			},
		},
		[]byte(`{"hash":"foo"}`),
		[][]byte{
			[]byte(`{"hash":"6a65e3082b44c5da7fa58a5c335b2a95acab3a925c7fa0cfd5bd6779ee7c2374"}`),
		},
		nil,
	},
	{
		"data md5",
		config.Config{
			Type: "proc_hash",
			Settings: map[string]interface{}{
				"algorithm": "md5",
			},
		},
		[]byte(`foo`),
		[][]byte{
			[]byte(`acbd18db4cc2f85cedef654fccc4a4d8`),
		},
		nil,
	},
	{
		"data sha256",
		config.Config{
			Type: "proc_hash",
			Settings: map[string]interface{}{
				"algorithm": "sha256",
			},
		},
		[]byte(`foo`),
		[][]byte{
			[]byte(`2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae`),
		},
		nil,
	},
}

func TestProcHash(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procHashTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcHash(ctx, test.cfg)
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

func benchmarkProcHash(b *testing.B, tform *procHash, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcHash(b *testing.B) {
	for _, test := range procHashTests {
		proc, err := newProcHash(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcHash(b, proc, test.test)
			},
		)
	}
}
