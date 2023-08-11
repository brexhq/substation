package transform_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

func ExampleTransformer() {
	ctx := context.TODO()

	// Copies the value of key "a" into key "b".
	cfg := config.Config{
		Type: "proc_copy",
		Settings: map[string]interface{}{
			"key":     "a",
			"set_key": "b",
		},
	}

	tf, err := transform.New(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// Transformer is applied to a message.
	message, err := mess.New(
		mess.SetData([]byte(`{"a":1}`)),
	)
	if err != nil {
		// handle err
		panic(err)
	}

	results, err := tf.Transform(ctx, message)
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
