package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

func main() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define a configuration. For native Substation applications, this is managed by Jsonnet.
	//
	// This example config copies a value and prints the data to stdout.
	conf := []byte(`
		{
			"transforms":[
				{"type":"proc_copy","settings":{"key":"a","set_key":"c"}},
				{"type":"send_stdout"}
			]
		}
	`)

	cfg := substation.Config{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		// Handle error.
		panic(err)
	}

	// Create a new Substation instance.
	sub, err := substation.New(ctx, cfg)
	if err != nil {
		// Handle error.
		panic(err)
	}

	// Always close the Substation instance.
	defer sub.Close(ctx)

	// Print the Substation configuration.
	fmt.Println(sub)

	// Substation instances process data defined as a Message. Messages can be processed
	// individually or in groups. This example processes multiple messages as a group.
	var msgs []*message.Message

	// Create a data Message and append it to the group.
	data, err := message.New(
		message.SetData([]byte(`{"a":"b"}`)),
	)
	if err != nil {
		// Handle error.
		panic(err)
	}

	msgs = append(msgs, data)

	// Create a control Message and append it to the group. Control messages trigger
	// special behavior in transforms. For example, a control message may be used
	// to flush buffered data or write a file to disk.
	ctrl, err := message.New(
		message.AsControl(),
	)
	if err != nil {
		// Handle error.
		panic(err)
	}

	msgs = append(msgs, ctrl)

	// Transform the group of messages. In this example, results are discarded.
	if _, err := transform.Apply(ctx, sub.Transforms(), msgs...); err != nil {
		// Handle error.
		panic(err)
	}
}
