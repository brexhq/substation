package substation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
	"github.com/brexhq/substation/v2/transform"
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
				{"type":"object_copy","settings":{"object":{"source_key":"a","target_key":"c"}}},
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
		message.New().AsControl(),
	}

	// Transform the group of messages. In this example, results are not used.
	if _, err := sub.Transform(ctx, msg...); err != nil {
		// Handle error.
		panic(err)
	}

	// Output:
	// {"transforms":[{"type":"object_copy","settings":{"object":{"source_key":"a","target_key":"c"}}},{"type":"send_stdout","settings":null}]}
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

func Example_substationCustomConfig() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define and load the custom configuration. This config includes a username
	// and password for authentication.
	conf := []byte(`
		{
			"transforms":[
				{"type":"object_copy","settings":{"object":{"source_key":"a","target_key":"c"}}},
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
	// {"transforms":[{"type":"object_copy","settings":{"object":{"source_key":"a","target_key":"c"}}},{"type":"send_stdout","settings":null}]}
	// foo:bar
}

func Example_substationCustomTransforms() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define and load the configuration. This config includes a transform that
	// is not part of the standard Substation package.
	conf := []byte(`
		{
			"transforms":[
				{"type":"utility_duplicate"},
				{"type":"send_stdout"}
			]
		}
	`)

	cfg := substation.Config{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		// Handle error.
		panic(err)
	}

	// Create a new Substation instance with a custom transform factory for loading
	// the custom transform.
	sub, err := substation.New(ctx, cfg, substation.WithTransformFactory(customFactory))
	if err != nil {
		// Handle error.
		panic(err)
	}

	msg := []*message.Message{
		message.New().SetData([]byte(`{"a":"b"}`)),
		message.New().AsControl(),
	}

	// Transform the group of messages. In this example, results are not used.
	if _, err := sub.Transform(ctx, msg...); err != nil {
		// Handle error.
		panic(err)
	}

	// Output:
	// {"a":"b"}
	// {"a":"b"}
}

// customFactory is used in the custom transform example to load the custom transform.
func customFactory(ctx context.Context, cfg config.Config) (transform.Transformer, error) {
	switch cfg.Type {
	// Usually a custom transform requires configuration, but this
	// is a toy example. Customizable transforms should have a new
	// function that returns a new instance of the configured transform.
	case "utility_duplicate":
		return &utilityDuplicate{Count: 1}, nil
	}

	return transform.New(ctx, cfg)
}

// Duplicates a message.
type utilityDuplicate struct {
	// Count is the number of times to duplicate the message.
	Count int `json:"count"`
}

func (t *utilityDuplicate) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Always return control messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	output := []*message.Message{msg}
	for i := 0; i < t.Count; i++ {
		output = append(output, msg)
	}

	return output, nil
}

func FuzzTestSubstation(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"transforms":[{"type":"utility_duplicate"}]}`),
		[]byte(`{"transforms":[{"type":"utility_duplicate", "count":2}]}`),
		[]byte(`{"transforms":[{"type":"unknown_type"}]}`),
		[]byte(`{"transforms":[{"type":"utility_duplicate", "count":"invalid"}]}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		var cfg substation.Config
		err := json.Unmarshal(data, &cfg)
		if err != nil {
			return
		}

		sub, err := substation.New(ctx, cfg)
		if err != nil {
			return
		}

		msg := message.New().SetData(data)
		_, err = sub.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
