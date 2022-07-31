# config
Contains functions for loading Substation data processing configurations. Substation includes [Jsonnet](https://jsonnet.org/) functions (`*.libsonnet`) to ease the burden of managing large, complex JSON configurations. Jsonnet examples can be found in the [configs/](configs/) directory.

The package can be used like this:
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

func main() {
	cfg := []byte(`{
		"settings": {
			"output_key": "foo",
			"options": {
				"value": "bar"
			}
		},
		"type": "insert"
	 }
	`)

	// unmarshal JSON object into Substation config
	var sub config.Config
	err := json.Unmarshal(cfg, &sub)
	if err != nil {
		panic(err)
	}

	// retrieve byter from the factory
	byter, err := process.ByterFactory(sub)
	if err != nil {
		panic(err)
	}

	// creates a new JSON object by inserting "bar" into key "foo"
	data, err := byter.Byte(context.TODO(), byte{})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
```
