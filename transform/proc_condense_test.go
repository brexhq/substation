package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
	"golang.org/x/exp/slices"
)

var _ Transformer = &procCondense{}

var procCondenseDataTests = []struct {
	name     string
	cfg      config.Config
	data     []string
	expected []string
}{
	{
		"no limit",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"separator": `\n`,
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
		[]string{
			`{"foo":"bar"}\n{"baz":"qux"}\n{"quux":"corge"}`,
		},
	},
	{
		"max_count",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"separator": `\n`,
				"max_count": 2,
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
		[]string{
			`{"foo":"bar"}\n{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
	},
	{
		"max_size",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"separator": `\n`,
				"max_size":  35,
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
		[]string{
			`{"foo":"bar"}\n{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
	},
	{
		"max_count and max_size",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"separator": `\n`,
				"max_count": 2,
				"max_size":  100,
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
		[]string{
			`{"foo":"bar"}\n{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
	},
	{
		"object array",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"separator": `\n`,
				"set_key":   "condense.-1",
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"baz":"qux"}`,
			`{"quux":"corge"}`,
		},
		[]string{
			`{"condense":[{"foo":"bar"},{"baz":"qux"},{"quux":"corge"}]}`,
		},
	},
}

func TestProcCondenseData(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procCondenseDataTests {
		t.Run(test.name, func(t *testing.T) {
			var messages []*mess.Message
			for _, data := range test.data {
				msg, err := mess.New(
					mess.SetData([]byte(data)),
				)
				if err != nil {
					t.Fatal(err)
				}

				messages = append(messages, msg)
			}

			// procCondense relies on a control message to flush the buffer,
			// so it's always added and then removed from the output.
			ctrl, err := mess.New(
				mess.AsControl(),
			)
			if err != nil {
				t.Fatal(err)
			}

			messages = append(messages, ctrl)

			proc, err := newProcCondense(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, messages...)
			if err != nil {
				t.Error(err)
			}

			var r []string
			for _, c := range result {
				if c.IsControl() {
					continue
				}

				r = append(r, string(c.Data()))
			}

			if !reflect.DeepEqual(r, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, r)
			}
		})
	}
}

func benchmarkProcCodenseData(b *testing.B, tform *procCondense, data []string) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		var messages []*mess.Message
		for _, d := range data {
			msg, _ := mess.New(
				mess.SetData([]byte(d)),
			)
			messages = append(messages, msg)
		}

		_, _ = tform.Transform(ctx, messages...)
	}
}

func BenchmarkProcCondenseData(b *testing.B) {
	for _, test := range procCondenseDataTests {
		proc, err := newProcCondense(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcCodenseData(b, proc, test.data)
			},
		)
	}
}

var procCondenseObjectTests = []struct {
	name     string
	cfg      config.Config
	data     []string
	expected []string
}{
	{
		"no limit",
		config.Config{
			Type: "condense",
			Settings: map[string]interface{}{
				"set_key":      "condense.-1",
				"condense_key": "foo",
			},
		},
		[]string{
			`{"foo":"bar"}`,
			`{"foo":"baz"}`,
			`{"foo":"bar"}`,
			`{"foo":"qux"}`,
			`{"foo":"bar"}`,
		},
		[]string{
			`{"condense":[{"foo":"bar"},{"foo":"bar"},{"foo":"bar"}]}`,
			`{"condense":[{"foo":"baz"}]}`,
			`{"condense":[{"foo":"qux"}]}`,
		},
	},
}

func TestProcCondenseObject(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procCondenseObjectTests {
		t.Run(test.name, func(t *testing.T) {
			var messages []*mess.Message
			for _, data := range test.data {
				msg, err := mess.New(
					mess.SetData([]byte(data)),
				)
				if err != nil {
					t.Fatal(err)
				}

				messages = append(messages, msg)
			}

			// procCondense relies on a control message to flush the buffer,
			// so it's always added and then removed from the output.
			ctrl, err := mess.New(
				mess.AsControl(),
			)
			if err != nil {
				t.Fatal(err)
			}

			messages = append(messages, ctrl)

			proc, err := newProcCondense(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, messages...)
			if err != nil {
				t.Error(err)
			}

			var arr []string
			for _, c := range result {
				if c.IsControl() {
					continue
				}

				arr = append(arr, string(c.Data()))
			}

			// The order of the output is not guaranteed, so we need to
			// check that the expected values are present anywhere in the
			// result.
			for _, r := range arr {
				if !slices.Contains(test.expected, r) {
					t.Errorf("expected %s, got %s", test.expected, r)
				}
			}
		})
	}
}
