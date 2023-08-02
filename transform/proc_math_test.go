package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procMath{}

var procMathTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"add",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "math",
				"set_key":   "math",
				"operation": "add",
			},
		},
		[]byte(`{"math":[1,3]}`),
		[][]byte{
			[]byte(`{"math":4}`),
		},
		nil,
	},
	{
		"subtract",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "math",
				"set_key":   "math",
				"operation": "subtract",
			},
		},
		[]byte(`{"math":[5,2]}`),
		[][]byte{
			[]byte(`{"math":3}`),
		},
		nil,
	},
	{
		"multiply",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "math",
				"set_key":   "math",
				"operation": "multiply",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[][]byte{
			[]byte(`{"math":20}`),
		},
		nil,
	},
	{
		"divide",
		config.Config{
			Settings: map[string]interface{}{
				"key":       "math",
				"set_key":   "math",
				"operation": "divide",
			},
		},
		[]byte(`{"math":[10,2]}`),
		[][]byte{
			[]byte(`{"math":5}`),
		},
		nil,
	},
}

func TestProcMath(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procMathTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcMath(ctx, test.cfg)
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

func benchmarkProcMath(b *testing.B, tform *procMath, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcMath(b *testing.B) {
	for _, test := range procMathTests {
		proc, err := newProcMath(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcMath(b, proc, test.test)
			},
		)
	}
}
