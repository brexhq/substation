package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procCase{}

var procCaseTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON lower",
		config.Config{
			Type: "proc_case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "lower",
			},
		},
		[]byte(`{"foo":"BAR"}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"JSON upper",
		config.Config{
			Type: "proc_case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "upper",
			},
		},
		[]byte(`{"foo":"bar"}`),
		[][]byte{
			[]byte(`{"foo":"BAR"}`),
		},
		nil,
	},
	{
		"JSON snake",
		config.Config{
			Type: "proc_case",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "snake",
			},
		},
		[]byte(`{"foo":"AbC"})`),
		[][]byte{
			[]byte(`{"foo":"ab_c"})`),
		},
		nil,
	},
}

func TestCase(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procCaseTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcCase(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
			if err != nil {
				t.Error(err)
			}

			var r [][]byte
			for _, c := range result {
				r = append(r, c.Data())
			}

			if !reflect.DeepEqual(r, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, r)
			}
		})
	}
}

func benchmarkProcCase(b *testing.B, tform *procCase, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, err := mess.New(
			mess.SetData(data),
		)
		if err != nil {
			b.Fatal(err)
		}

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcCase(b *testing.B) {
	for _, test := range procCaseTests {
		proc, err := newProcCase(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcCase(b, proc, test.test)
			},
		)
	}
}
