package substation_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
)

func ExampleSubstation() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define a configuration. For native Substation applications, this is managed by Jsonnet.
	//
	// This example copies an object's value and prints the data to stdout.
	conf := []byte(`
		{
			"transforms":[
				{"type":"object_copy","settings":{"object":{"key":"a","set_key":"c"}}},
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

	// Print the Substation configuration.
	fmt.Println(sub)

	// Substation instances process data defined as a Message. Messages can be processed
	// individually or in groups. This example processes multiple messages as a group.
	msg := []*message.Message{
		// The first message is a data message. Only data messages are transformed.
		message.New().SetData([]byte(`{"a":"b"}`)),
		// The second message is a ctrl message. ctrl messages flush the pipeline.
		message.New(message.AsControl()),
	}

	// Transform the group of messages. In this example, results are not used.
	if _, err := sub.Transform(ctx, msg...); err != nil {
		// Handle error.
		panic(err)
	}

	// Output:
	// {"transforms":[{"type":"object_copy","settings":{"object":{"key":"a","set_key":"c"}}},{"type":"send_stdout","settings":null}]}
	// {"a":"b","c":"b"}
}

// Custom applications should embed the Substation configuration and
// add additional configuration options.
type customConfig struct {
	substation.Config

	Auth struct {
		Username string `json:"username"`
		// Please don't store passwords in configuration files, this is only an example!
		Password string `json:"password"`
	} `json:"auth"`
}

// String returns an example string representation of the custom configuration.
func (c customConfig) String() string {
	return fmt.Sprintf("%s:%s", c.Auth.Username, c.Auth.Password)
}

func Example_customSubstation() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define and load the custom configuration. This config includes a username
	// and password for authentication.
	conf := []byte(`
		{
			"transforms":[
				{"type":"object_copy","settings":{"object":{"key":"a","set_key":"c"}}},
				{"type":"send_stdout"}
			],
			"auth":{
				"username":"foo",
				"password":"bar"
			}
		}
	`)

	cfg := customConfig{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		// Handle error.
		panic(err)
	}

	// Create a new Substation instance from the embedded configuration.
	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		// Handle error.
		panic(err)
	}

	// Print the Substation configuration.
	fmt.Println(sub)

	// Print the custom configuration.
	fmt.Println(cfg)

	// Output:
	// {"transforms":[{"type":"object_copy","settings":{"object":{"key":"a","set_key":"c"}}},{"type":"send_stdout","settings":null}]}
	// foo:bar
}
