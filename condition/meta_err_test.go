package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &metaErr{}

var metaErrTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"catch_all",
		config.Config{
			Settings: map[string]interface{}{
				"inspector": map[string]interface{}{
					"settings": map[string]interface{}{
						"object": map[string]interface{}{
							"source_key": "a",
						},
						"inspector": map[string]interface{}{
							"type": "string_starts_with",
							"settings": map[string]interface{}{
								"string": "c",
							},
						},
						"type": "any",
					},
					"type": "meta_for_each",
				},
			},
		},
		[]byte(`{"a":"bcd"}`),
		false,
	},
	{
		"catch_one",
		config.Config{
			Settings: map[string]interface{}{
				"error_messages": []string{"input must be an array"},
				"inspector": map[string]interface{}{
					"settings": map[string]interface{}{
						"object": map[string]interface{}{
							"source_key": "a",
						},
						"inspector": map[string]interface{}{
							"type": "string_starts_with",
							"settings": map[string]interface{}{
								"string": "c",
							},
						},
						"type": "any",
					},
					"type": "meta_for_each",
				},
			},
		},
		[]byte(`{"a":"bcd"}`),
		false,
	},
	{
		"no_error",
		config.Config{
			Settings: map[string]interface{}{
				"inspector": map[string]interface{}{
					"settings": map[string]interface{}{
						"object": map[string]interface{}{
							"source_key": "a",
						},
						"inspector": map[string]interface{}{
							"type": "string_starts_with",
							"settings": map[string]interface{}{
								"string": "c",
							},
						},
						"type": "any",
					},
					"type": "meta_for_each",
				},
			},
		},
		[]byte(`{"a":["bcd"]}`),
		true,
	},
}

func TestMetaErr(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaErrTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaErr(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.data))
			}
		})
	}
}

func benchmarkMetaErr(b *testing.B, insp *metaErr, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkMetaErr(b *testing.B) {
	for _, test := range metaErrTests {
		insp, err := newMetaErr(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkMetaErr(b, insp, message)
			},
		)
	}
}
