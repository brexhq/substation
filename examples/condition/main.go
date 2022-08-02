package main

import (
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

func main() {
	cfg := []byte(`{
		"settings": {
			"key": "foo",
			"expression": "bar",
			"function": "equals"
		},
		"type": "strings"
	 }`)

	// unmarshal JSON object into Substation config
	var sub config.Config
	err := json.Unmarshal(cfg, &sub)
	if err != nil {
		panic(err)
	}

	// retrieve inspector from the factory
	inspector, err := condition.InspectorFactory(sub)
	if err != nil {
		panic(err)
	}

	data := []byte(`{"foo":"bar"}`)
	ok, err := inspector.Inspect(data)
	if err != nil {
		panic(err)
	}

	if ok {
		fmt.Println("data passed inspection")
	} else {
		fmt.Println("data failed inspection")
	}
}
