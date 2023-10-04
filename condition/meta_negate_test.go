package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &metaNegate{}

var metaNegateTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"inspector": map[string]interface{}{
					"settings": map[string]interface{}{
						"object": map[string]interface{}{
							"key": "a",
						},
						"string": "bcd",
					},
					"type": "string_equal_to",
				},
			},
		},
		[]byte(`{"a":"bcd"}`),
		false,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"inspector": map[string]interface{}{
					"type": "string_equal_to",
					"settings": map[string]interface{}{
						"string": "bcd",
					},
				},
			},
		},
		[]byte(`bcd`),
		false,
	},
}

func TestMetaNegate(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaNegateTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaNegate(ctx, test.cfg)
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

func benchmarkMetaNegate(b *testing.B, insp *metaNegate, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkMetaNegate(b *testing.B) {
	for _, test := range metaNegateTests {
		insp, err := newMetaNegate(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkMetaNegate(b, insp, message)
			},
		)
	}
}
