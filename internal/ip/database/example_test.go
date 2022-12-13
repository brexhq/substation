package database_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/internal/ip/database"
)

func Example_iP2Location() {
	// create IP2Location container, open database, and close database when function returns
	ip2loc := database.IP2Location{
		Database: "location://path/to/ip2location.bin",
	}

	if err := ip2loc.Open(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
	defer ip2loc.Close()

	// query database
	addr := "8.8.8.8"
	record, err := ip2loc.Get(addr)
	if err != nil {
		// handle error
		panic(err)
	}

	// marshal to JSON for printing
	res, _ := json.Marshal(record)
	fmt.Println(string(res))
}

func Example_maxMindCity() {
	// create MaxMind City container, open database, and close database when function returns
	mm := database.MaxMindCity{
		Database: "location://path/to/maxmind.mmdb",
	}

	if err := mm.Open(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
	defer mm.Close()

	// query database
	addr := "8.8.8.8"
	record, err := mm.Get(addr)
	if err != nil {
		// handle error
		panic(err)
	}

	// marshal to JSON for printing
	res, _ := json.Marshal(record)
	fmt.Println(string(res))
}
