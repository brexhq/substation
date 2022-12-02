package process_test

import (
	"context"
	"fmt"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

func Example_iPInfo() {
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"ip":"8.8.8.8"}`))

	// the location of the IP enrichment database must be provided by environment variable and can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
	//nolint:tenv // example doesn't use testing package
	_ = os.Setenv("MAXMIND_ASN_DB", "location://path/to/maxmind.mmdb")
	defer os.Unsetenv("MAXMIND_ASN_DB")

	//nolint:tenv // example doesn't use testing package
	_ = os.Setenv("MAXMIND_LOCATION_DB", "location://path/to/maxmind.mmdb")
	defer os.Unsetenv("MAXMIND_LOCATION_DB")

	// in native Substation applications configuration is handled by compiling Jsonnet and loading JSON into the application
	cfg := []config.Config{
		{
			Type: "ip_info",
			Settings: map[string]interface{}{
				"input_key":  "ip",
				"output_key": "as",
				"options": map[string]interface{}{
					"function": "maxmind_asn",
				},
			},
		},
		{
			Type: "ip_info",
			Settings: map[string]interface{}{
				"input_key":  "ip",
				"output_key": "geo",
				"options": map[string]interface{}{
					"function": "maxmind_geo",
				},
			},
		},
	}

	applicators, err := process.MakeApplicators(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, applicator := range applicators {
		// applies the IPInfo processors to the capsule
		capsule, err = applicator.Apply(context.TODO(), capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	fmt.Println(string(capsule.Data()))
}
