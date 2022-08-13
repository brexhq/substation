// example of reading data from a file and applying a
// single processor to a batch of data
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
	// read data file into slice of encapsulated data
	open, err := os.Open("./data.json")
	if err != nil {
		panic(err)
	}

	var caps []config.Capsule
	cap := config.NewCapsule()

	scanner := bufio.NewScanner(open)
	for scanner.Scan() {
		cap.SetData(scanner.Bytes())
		caps = append(caps, cap)
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

	proc, err := process.BatchFactory(sub)
	if err != nil {
		panic(err)
	}

	// apply batch processor to encapsulated data
	caps, err = process.ApplyBatch(context.TODO(), caps, proc)
	if err != nil {
		panic(err)
	}

	for _, cap := range caps {
		fmt.Printf("%s\n", cap.GetData())
	}
}
