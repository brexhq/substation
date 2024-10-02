package transform

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func FuzzTestMetaMetricDuration(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"c":"d"}`),
		[]byte(`{"e":"f"}`),
		[]byte(`{"a":{"b":"c"}}`),
		[]byte(`{"array":[1,2,3]}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newMetaMetricsDuration(ctx, config.Config{
			Settings: map[string]interface{}{
				"transforms": []config.Config{
					{
						Type: "format_from_base64",
					},
				},
				"metric": map[string]interface{}{
					"destination": "stdout",
				},
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
