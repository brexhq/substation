package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procGzip{}

var procGzipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"from",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[][]byte{
			[]byte(`foo`),
		},
		nil,
	},
	{
		"to",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "to",
			},
		},
		[]byte(`foo`),
		[][]byte{
			{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		},
		nil,
	},
}

func TestProcGzip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procGzipTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcGzip(ctx, test.cfg)
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

func benchmarkProcGzip(b *testing.B, tform *procGzip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcGzip(b *testing.B) {
	for _, test := range procGzipTests {
		proc, err := newProcGzip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcGzip(b, proc, test.test)
			},
		)
	}
}
