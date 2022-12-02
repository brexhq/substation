package ip_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/internal/ip"
)

func Example_iP2Location() {
	// the location of the IP enrichment database must be provided by environment variable and can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
	//nolint:tenv // example doesn't use testing package
	_ = os.Setenv("IP2LOCATION_DB", "location://path/to/ip2location.bin")
	defer os.Unsetenv("IP2LOCATION_DB")

	// create IP2Location container, load database, and close database when function returns
	ip2loc := ip.IP2Location{}
	if err := ip2loc.Setup(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
	defer ip2loc.Close()

	// lookup location information
	addr := "8.8.8.8"
	location, err := ip2loc.Location(addr)
	if err != nil {
		// handle error
		panic(err)
	}

	// marshal to JSON for printing
	res, _ := json.Marshal(location)
	fmt.Println(string(res))
}

func Example_maxMind() {
	// the location of the IP enrichment database must be provided by environment variable and can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
	//nolint:tenv // example doesn't use testing package
	_ = os.Setenv("MAXMIND_LOCATION_DB", "location://path/to/maxmind.mmdb")
	defer os.Unsetenv("MAXMIND_LOCATION_DB")

	// create MaxMind container, load database, and close database when function returns
	mm := ip.MaxMind{}
	if err := mm.Setup(context.TODO()); err != nil {
		// handle error
		panic(err)
	}
	defer mm.Close()

	// lookup location information
	addr := "8.8.8.8"
	location, err := mm.Location(addr)
	if err != nil {
		// handle error
		panic(err)
	}

	// marshal to JSON for printing
	res, _ := json.Marshal(location)
	fmt.Println(string(res))
}
