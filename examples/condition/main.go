// example of reading data from a file and applying a inspector
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

func main() {
	cfg, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	// unmarshal JSON object into Substation config
	var sub config.Config
	if err := json.Unmarshal(cfg, &sub); err != nil {
		panic(err)
	}

	// retrieve inspector from the factory
	inspector, err := condition.InspectorFactory(sub)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile("./data.json")
	if err != nil {
		panic(err)
	}

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
