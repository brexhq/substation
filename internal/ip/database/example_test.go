package database_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/internal/ip/database"
)

func Example_iP2Location() {
	// the location of the enrichment database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
	location := "location://path/to/ip2location.bin"

	// create IP2Location container, open database, and close database when function returns
	ip2loc := database.IP2Location{}
	if err := ip2loc.Open(context.TODO(), location); err != nil {
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
	// the location of the enrichment database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
	location := "location://path/to/maxmind.mmdb"

	// create MaxMind City container, open database, and close database when function returns
	mm := database.MaxMindCity{}
	if err := mm.Open(context.TODO(), location); err != nil {
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
