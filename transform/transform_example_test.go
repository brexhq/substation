package transform_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
	"github.com/brexhq/substation/v2/transform"
)

func ExampleTransformer() {
	ctx := context.TODO()

	// Copies the value of key "a" into key "b".
	cfg := config.Config{
		Type: "object_copy",
		Settings: map[string]interface{}{
			"object": map[string]interface{}{
				"source_key": "a",
				"target_key": "b",
			},
		},
	}

	tf, err := transform.New(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// Transformer is applied to a message.
	msg := message.New().SetData([]byte(`{"a":1}`))
	results, err := tf.Transform(ctx, msg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, c := range results {
		fmt.Println(string(c.Data()))
	}

	// Output:
	// {"a":1,"b":1}
}
