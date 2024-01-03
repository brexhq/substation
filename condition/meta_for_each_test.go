package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ inspector = &metaForEach{}

var metaForEachTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
	err      error
}{
	{
		"string starts_with all",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
				},
				"type": "all",
				"inspector": map[string]interface{}{
					"type": "string_starts_with",
					"settings": map[string]interface{}{
						"string": "f",
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"ip private all",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
				},
				"type": "all",
				"inspector": map[string]interface{}{
					"type": "network_ip_private",
				},
			},
		},
		[]byte(`{"input":["192.168.1.2","10.0.42.1","172.16.4.2"]}`),
		true,
		nil,
	},
	{
		"string_match",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
				},
				"type": "any",
				"inspector": map[string]interface{}{
					"type": "string_match",
					"settings": map[string]interface{}{
						"pattern": "^fizz$",
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"string greater_than",
		config.Config{
			Settings: map[string]interface{}{
				"type": "any",
				"inspector": map[string]interface{}{
					"type": "string_greater_than",
					"settings": map[string]interface{}{
						"string": "0",
					},
				},
			},
		},
		[]byte(`[0,1,2]`),
		true,
		nil,
	},
	{
		"length none",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
				},
				"type": "none",
				"inspector": map[string]interface{}{
					"type": "number_length_greater_than",
					"settings": map[string]interface{}{
						"value": 7,
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"length all",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
				},
				"type": "all",
				"inspector": map[string]interface{}{
					"type": "number_length_equal_to",
					"settings": map[string]interface{}{
						"value": 4,
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
}

func TestMetaForEach(t *testing.T) {
	ctx := context.TODO()

	for _, tt := range metaForEachTests {
		t.Run(tt.name, func(t *testing.T) {
			message := message.New().SetData(tt.test)

			insp, err := newMetaForEach(ctx, tt.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if tt.expected != check {
				t.Errorf("expected %v, got %v, %v", tt.expected, check, string(tt.test))
			}
		})
	}
}

func benchmarkMetaForEach(b *testing.B, insp *metaForEach, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkMetaForEach(b *testing.B) {
	for _, test := range metaForEachTests {
		insp, err := newMetaForEach(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkMetaForEach(b, insp, message)
			},
		)
	}
}
