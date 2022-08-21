// example of reading data from a file and applying an inspector
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
)

func main() {
	// read lines from data file
	open, err := os.Open("../data.json")
	if err != nil {
		panic(err)
	}

	var data [][]byte

	scanner := bufio.NewScanner(open)
	for scanner.Scan() {
		data = append(data, scanner.Bytes())
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
	for _, data := range data {
		ok, err := condition.InspectByte(data, inspector)
		if err != nil {
			panic(err)
		}

		if ok {
			fmt.Printf("passed inspection: %s\n", data)
		} else {
			fmt.Printf("failed inspection: %s\n", data)
		}
	}
}
