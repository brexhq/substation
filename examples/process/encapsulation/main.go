// example of reading data from a file and applying a
// single processor to encapsulated data
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

func main() {
	// read lines from data file into encapsulated data
	open, err := os.Open("../data.json")
	if err != nil {
		panic(err)
	}

	var capsules []config.Capsule
	capsule := config.NewCapsule()

	scanner := bufio.NewScanner(open)
	for scanner.Scan() {
		capsule.SetData(scanner.Bytes())
		capsules = append(capsules, capsule)
	}

	// read config file and create a new batch processor
	cfg, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var sub config.Config
	if err := json.Unmarshal(cfg, &sub); err != nil {
		panic(err)
	}

	batchers, err := process.MakeBatchers(sub)
	if err != nil {
		panic(err)
	}

	// apply batch processor to encapsulated data
	capsules, err = process.Batch(context.TODO(), capsules, batchers...)
	if err != nil {
		panic(err)
	}

	for _, capsule := range capsules {
		fmt.Printf("%s\n", capsule.Data())
	}
}
