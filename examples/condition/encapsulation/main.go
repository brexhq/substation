// example of reading data from a file and applying an inspector
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

func main() {
	// read lines from data file as encapsulated data
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

	// read config file and create a new inspector
	cfg, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var sub config.Config
	if err := json.Unmarshal(cfg, &sub); err != nil {
		panic(err)
	}

	inspector, err := condition.InspectorFactory(sub)
	if err != nil {
		panic(err)
	}

	// apply inspector to encapsulated data
	for _, capsule := range capsules {
		ok, err := inspector.Inspect(context.TODO(), capsule)
		if err != nil {
			panic(err)
		}

		if ok {
			fmt.Printf("passed inspection: %s\n", capsule.Data())
		} else {
			fmt.Printf("failed inspection: %s\n", capsule.Data())
		}
	}
}
