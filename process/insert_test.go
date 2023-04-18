package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procInsert{}
	_ Batcher = procInsert{}
)

var insertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"string",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": "foo",
				},
			},
		},
		[]byte{},
		[]byte(`{"insert":"foo"}`),
		nil,
	},
	{
		"int",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": 10,
				},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":10}`),
		nil,
	},
	{
		"string array",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": []string{"bar", "baz"},
				},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":["bar","baz"]}`),
		nil,
	},
	{
		"map",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": map[string]string{
						"bar": "baz",
					},
				},
			},
		},
		[]byte(`{"insert":"foo"}`),
		[]byte(`{"insert":{"bar":"baz"}}`),
		nil,
	},
	{
		"JSON",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": `{"bar":"baz"}`,
				},
			},
		},
		[]byte(`{"insert":"bar"}`),
		[]byte(`{"insert":{"bar":"baz"}}`),
		nil,
	},
	{
		"zlib",
		config.Config{
			Type: "insert",
			Settings: map[string]interface{}{
				"set_key": "insert",
				"options": map[string]interface{}{
					"value": []byte{120, 156, 5, 192, 49, 13, 0, 0, 0, 194, 48, 173, 76, 2, 254, 143, 166, 29, 2, 93, 1, 54},
				},
			},
		},
		[]byte(`{"insert":"bar"}`),
		[]byte(`{"insert":"eJwFwDENAAAAwjCtTAL+j6YdAl0BNg=="}`),
		nil,
	},
}

func TestInsert(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range insertTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcInsert(ctx, test.cfg)
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

func benchmarkInsert(b *testing.B, applier procInsert, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkInsert(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range insertTests {
		proc, err := newProcInsert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkInsert(b, proc, capsule)
			},
		)
	}
}
