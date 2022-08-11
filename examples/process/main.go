// example of reading data from a file and applying a single processor
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
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

	// retrieve byter from the factory
	byter, err := process.ByterFactory(sub)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile("./data.json")
	if err != nil {
		panic(err)
	}

	// inserts "qux" into key "baz"
	data, err = byter.Byte(context.TODO(), data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
