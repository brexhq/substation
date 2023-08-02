package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Inspector = &metaInspCondition{}

var metaConditionTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"object",
		config.Config{
			Type: "meta_condition",
			Settings: map[string]interface{}{
				"condition": Config{
					Operator: "all",
					Inspectors: []config.Config{
						{
							Type: "insp_ip",
							Settings: map[string]interface{}{
								"key":  "ip_address",
								"type": "private",
							},
						},
					},
				},
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"data",
		config.Config{
			Type: "meta_condition",
			Settings: map[string]interface{}{
				"condition": Config{
					Operator: "all",
					Inspectors: []config.Config{
						{
							Type: "insp_ip",
							Settings: map[string]interface{}{
								"type": "private",
							},
						},
					},
				},
			},
		},
		[]byte("192.168.1.2"),
		true,
	},
}

func TestMetaCondition(t *testing.T) {
	ctx := context.TODO()

	for _, test := range metaConditionTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.data),
			)

			insp, err := newMetaInspCondition(ctx, test.cfg)
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

func benchmarkMetaCondition(b *testing.B, inspector *metaInspCondition, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkMetaCondition(b *testing.B) {
	for _, test := range metaConditionTests {
		insp, err := newMetaInspCondition(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.data),
				)
				benchmarkMetaCondition(b, insp, message)
			},
		)
	}
}
