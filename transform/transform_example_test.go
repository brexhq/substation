package transform_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

func ExampleNewTransformer() {
	ctx := context.TODO()

	// Copies the value of key "a" into key "b".
	cfg := config.Config{
		Type: "proc_copy",
		Settings: map[string]interface{}{
			"key":     "a",
			"set_key": "b",
		},
	}

	// One or more transforms can be created from a config.
	tform, err := transform.NewTransformer(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(tform)
	// Output: {"key":"a","set_key":"b"}
}

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

	// One or more transforms can be created from a config.
	tform, err := transform.NewTransformer(ctx, cfg)
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

	results, err := tform.Transform(ctx, message)
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

func Example_dNS() {
	ctx := context.TODO()

	message, err := mess.New(
		mess.SetData([]byte(`{"addr":"8.8.8.8"}`)),
	)
	if err != nil {
		// handle err
		panic(err)
	}

	// Process the message with the DNS transform.
	cfg := []config.Config{
		{
			Type: "dns",
			Settings: map[string]interface{}{
				"key":     "addr",
				"set_key": "domains",
				"options": map[string]interface{}{
					"function": "reverse_lookup",
				},
			},
		},
	}

	transforms, err := transform.NewTransformers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // Errors are ignored in case processing fails in a single transform.
	defer transform.CloseTransformers(ctx, transforms...)

	results, err := transform.Apply(ctx, transforms, message)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, c := range results {
		fmt.Println(string(c.Data()))
	}
}

func Example_hTTP() {
	ctx := context.TODO()
	message, err := mess.New(
		mess.SetData([]byte(`{"addr":"8.8.8.8"}`)),
	)
	if err != nil {
		// handle err
		panic(err)
	}

	// enriches the IP address by querying the GreyNoise Community API.
	// authenticating to GreyNoise is accomplished by interpolating a
	// secret inside an HTTP header. refer to the GreyNoise documentation
	// for more information:
	// https://docs.greynoise.io/reference/get_v3-community-it.
	cfg := []config.Config{
		{
			Type: "http",
			Settings: map[string]interface{}{
				"key": "addr",
				// the HTTP response body is written to this key
				"set_key": "greynoise",
				"options": map[string]interface{}{
					"method": "key",
					// the value from "addr" is interpolated into ${data}
					"url": "https://api.greynoise.io/v3/community/${data}",
					"headers": []map[string]interface{}{
						{
							"key": "key",
							// this secret must be stored in the environment
							// variable GREYNOISE_API
							"value": "${SECRETS_ENV:GREYNOISE_API}",
						},
					},
				},
			},
		},
	}

	transforms, err := transform.NewTransformers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // Errors are ignored in case processing fails in a single transform.
	defer transform.CloseTransformers(ctx, transforms...)

	results, err := transform.Apply(ctx, transforms, message)
	if err != nil {
		// handle err
		panic(err)
	}

	// sample output (which may change day to day)
	// {"addr":"8.8.8.8","greynoise":{"ip":"8.8.8.8","noise":false,"riot":true,"classification":"benign","name":"Google Public DNS","link":"https://viz.greynoise.io/riot/8.8.8.8","last_seen":"2023-01-30","message":"Success"}}
	for _, c := range results {
		fmt.Println(string(c.Data()))
	}
}

func Example_kVStore() {
	ctx := context.TODO()

	message, err := mess.New(
		mess.SetData([]byte(`{"a":"b"}`)),
	)
	if err != nil {
		// handle err
		panic(err)
	}

	// The value from key "a" is first set into the KV store and
	// then retrieved from the KV store and set into key "c". if
	// the KV options are identical across transforms, then the same
	// KV store is used in each call. This also allows for the use of
	// multiple KV stores.
	cfg := []config.Config{
		{
			Type: "kv_store",
			Settings: map[string]interface{}{
				"key":     "a",
				"set_key": "setter",
				"options": map[string]interface{}{
					"type": "set_key",
					"kv_options": map[string]interface{}{
						"type": "memory",
						"settings": map[string]interface{}{
							"capacitiy": 10,
						},
					},
				},
			},
		},
		{
			Type: "kv_store",
			Settings: map[string]interface{}{
				"key":     "setter",
				"set_key": "c",
				"options": map[string]interface{}{
					"type": "key",
					"kv_options": map[string]interface{}{
						"type": "memory",
						"settings": map[string]interface{}{
							"capacitiy": 10,
						},
					},
				},
			},
		},
	}

	transforms, err := transform.NewTransformers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // Errors are ignored in case processing fails in a single transform.
	defer transform.CloseTransformers(ctx, transforms...)

	results, err := transform.Apply(ctx, transforms, message)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, c := range results {
		fmt.Println(string(c.Data()))
	}
}
