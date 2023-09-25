package transform

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	mess "github.com/brexhq/substation/message"
	"golang.org/x/exp/slices"
)

var aggregateToStringTests = []struct {
	name     string
	cfg      config.Config
	data     []string
	expected []string
}{
	{
		"data no_limit",
		config.Config{
			Settings: map[string]interface{}{
				"separator": `\n`,
			},
		},
		[]string{
			`{"a":"b"}`,
			`{"c":"d"}`,
			`{"e":"f"}`,
		},
		[]string{
			`{"a":"b"}\n{"c":"d"}\n{"e":"f"}`,
		},
	},
	{
		"data with_key",
		config.Config{
			Settings: map[string]interface{}{
				"buffer": map[string]interface{}{
					"key": "c",
				},
				"separator": `\n`,
			},
		},
		[]string{
			`{"a":"b"}`,
			`{"c":"d"}`,
			`{"e":"f"}`,
		},
		[]string{
			`{"a":"b"}\n{"e":"f"}`,
			`{"c":"d"}`,
		},
	},
	{
		"data max_count",
		config.Config{
			Settings: map[string]interface{}{
				"buffer": map[string]interface{}{
					"count": 2,
				},
				"separator": `\n`,
			},
		},
		[]string{
			`{"a":"b"}`,
			`{"c":"d"}`,
			`{"e":"f"}`,
		},
		[]string{
			`{"a":"b"}\n{"c":"d"}`,
			`{"e":"f"}`,
		},
	},
	{
		"data max_size",
		config.Config{
			Settings: map[string]interface{}{
				"buffer": map[string]interface{}{
					"size": 25,
				},
				"separator": `\n`,
			},
		},
		[]string{
			`{"a":"b"}`,
			`{"c":"d"}`,
			`{"e":"f"}`,
		},
		[]string{
			`{"a":"b"}\n{"c":"d"}`,
			`{"e":"f"}`,
		},
	},
}

func TestAggregateToString(t *testing.T) {
	ctx := context.TODO()
	for _, test := range aggregateToStringTests {
		t.Run(test.name, func(t *testing.T) {
			var messages []*message.Message
			for _, data := range test.data {
				msg := message.New().SetData([]byte(data))
				messages = append(messages, msg)
			}

			// aggregateToString relies on an interrupt message to flush the buffer,
			// so it's always added and then removed from the output.
			ctrl := message.New(mess.AsControl())
			messages = append(messages, ctrl)

			tf, err := newAggregateToString(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := Apply(ctx, []Transformer{tf}, messages...)
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
