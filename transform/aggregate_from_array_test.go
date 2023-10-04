package transform

import (
	"context"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &aggregateFromArray{}

var aggregateFromArrayTests = []struct {
	name     string
	cfg      config.Config
	data     []string
	expected []string
}{
	// data tests
	{
		"data",
		config.Config{},
		[]string{
			`[{"a":"b"},{"c":"d"},{"e":"f"}]`,
		},
		[]string{
			`{"a":"b"}`,
			`{"c":"d"}`,
			`{"e":"f"}`,
		},
	},
	{
		"data with set_key",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"set_key": "x",
				},
			},
		},
		[]string{
			`[{"a":"b"},{"c":"d"},{"e":"f"}]`,
		},
		[]string{
			`{"x":{"a":"b"}}`,
			`{"x":{"c":"d"}}`,
			`{"x":{"e":"f"}}`,
		},
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "x",
				},
			},
		},
		[]string{
			`{"x":[{"a":"b"},{"c":"d"},{"e":"f"}],"y":"z"}`,
		},
		[]string{
			`{"y":"z","a":"b"}`,
			`{"y":"z","c":"d"}`,
			`{"y":"z","e":"f"}`,
		},
	},
	{
		"object with set_key",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "x",
					"set_key": "x",
				},
			},
		},
		[]string{
			`{"x":[{"a":"b"},{"c":"d"},{"e":"f"}],"y":"z"}`,
		},
		[]string{
			`{"y":"z","x":{"a":"b"}}`,
			`{"y":"z","x":{"c":"d"}}`,
			`{"y":"z","x":{"e":"f"}}`,
		},
	},
}

func TestAggregateFromArray(t *testing.T) {
	ctx := context.TODO()
	for _, test := range aggregateFromArrayTests {
		t.Run(test.name, func(t *testing.T) {
			var messages []*message.Message
			for _, data := range test.data {
				msg := message.New().SetData([]byte(data))
				messages = append(messages, msg)
			}

			// aggregateFromArray relies on an interrupt message to flush the buffer,
			// so it's always added and then removed from the output.
			ctrl := message.New(mess.AsControl())
			messages = append(messages, ctrl)

			tf, err := newAggregateFromArray(ctx, test.cfg)
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
