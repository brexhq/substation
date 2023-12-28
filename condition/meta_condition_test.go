package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &metaCondition{}

var metaConditionTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"condition": Config{
					Operator: "all",
					Inspectors: []config.Config{
						{
							Type: "string_contains",
							Settings: map[string]interface{}{
								"object": map[string]interface{}{
									"src_key": "a",
								},
								"string": "bcd",
							},
						},
					},
				},
			},
		},
		[]byte(`{"a":"bcd"}`),
		true,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"condition": Config{
					Operator: "all",
					Inspectors: []config.Config{
						{
							Type: "string_contains",
							Settings: map[string]interface{}{
								"string": "bcd",
							},
						},
					},
				},
			},
		},
		[]byte("bcd"),
		true,
	},
}

func TestMetaCondition(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaConditionTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newMetaCondition(ctx, test.cfg)
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

func benchmarkMetaCondition(b *testing.B, inspector *metaCondition, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkMetaCondition(b *testing.B) {
	for _, test := range metaConditionTests {
		insp, err := newMetaCondition(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkMetaCondition(b, insp, message)
			},
		)
	}
}
