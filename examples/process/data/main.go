// example of reading data from a file and applying a
// single processor to data
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
	// read lines from data file as encapsulated data
	open, err := os.Open("../data.json")
	if err != nil {
		panic(err)
	}

	var data [][]byte

	scanner := bufio.NewScanner(open)
	for scanner.Scan() {
		data = append(data, scanner.Bytes())
	}

	// read config file and create a new batch processor
	cfg, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var conf config.Config
	if err := json.Unmarshal(cfg, &conf); err != nil {
		panic(err)
	}

	applicators, err := process.MakeApplicators(conf)
	if err != nil {
		panic(err)
	}

	// apply processor to data
	for _, data := range data {
		data, err = process.ApplyBytes(context.TODO(), data, applicators...)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", data)
	}
}
