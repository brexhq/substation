package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation"
)

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

func main() {
	// Substation applications rely on a context for cancellation and timeouts.
	ctx := context.Background()

	// Define and load the custom configuration. This config includes a username
	// and password for authentication.
	conf := []byte(`
		{
			"transforms":[
				{"type":"proc_copy","settings":{"key":"a","set_key":"c"}},
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

	// Always close the Substation instance.
	defer sub.Close(ctx)

	// Print the Substation configuration.
	fmt.Println(sub)

	// Print the custom configuration.
	fmt.Println(cfg)
}
