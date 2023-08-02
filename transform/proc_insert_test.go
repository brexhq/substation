package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procInsert{}

var procInsertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"string",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value":   "foo",
			},
		},
		[]byte{},
		[][]byte{
			[]byte(`{"insert":"foo"}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value":   10,
			},
		},
		[]byte(`{"insert":"foo"}`),
		[][]byte{
			[]byte(`{"insert":10}`),
		},
		nil,
	},
	{
		"string array",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value":   []string{"bar", "baz"},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[][]byte{
			[]byte(`{"insert":["bar","baz"]}`),
		},
		nil,
	},
	{
		"map",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value": map[string]string{
					"bar": "baz",
				},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[][]byte{
			[]byte(`{"insert":{"bar":"baz"}}`),
		},
		nil,
	},
	{
		"JSON",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value":   `{"bar":"baz"}`,
			},
		},
		[]byte(`{"insert":"bar"}`),
		[][]byte{
			[]byte(`{"insert":{"bar":"baz"}}`),
		},
		nil,
	},
	{
		"zlib",
		config.Config{
			Type: "proc_insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"value":   []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
			},
		},
		[]byte(`{"insert":"bar"}`),
		[][]byte{
			[]byte(`{"insert":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		},
		nil,
	},
}

func TestProcInsert(t *testing.T) {
	ctx := context.TODO()

	for _, test := range procInsertTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcInsert(ctx, test.cfg)
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

func benchmarkProcInsert(b *testing.B, tform *procInsert, data []byte) {
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

func BenchmarkProcInsert(b *testing.B) {
	for _, test := range procInsertTests {
		proc, err := newProcInsert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcInsert(b, proc, test.test)
			},
		)
	}
}
