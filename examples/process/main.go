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
			"output_key": "baz",
			"options": {
				"value": "qux"
			}
		},
		"type": "insert"
	 }`)

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

	data := []byte(`{"foo":"bar"}`)
	fmt.Println(string(data))

	// inserts "qux" into key "baz"
	data, err = byter.Byte(context.TODO(), data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
