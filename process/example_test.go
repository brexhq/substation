package process_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

func Example_dNS() {
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"ip":"8.8.8.8"}`))

	// in native Substation applications configuration is handled by compiling Jsonnet and loading JSON into the application
	cfg := []config.Config{
		{
			Type: "dns",
			Settings: map[string]interface{}{
				"input_key":  "ip",
				"output_key": "domains",
				"options": map[string]interface{}{
					"function": "reverse_lookup",
				},
			},
		},
	}

	applicators, err := process.MakeApplicators(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applicator
	defer process.CloseApplicators(context.TODO(), applicators...)

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

func Example_iPDatabase() {
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"ip":"8.8.8.8"}`))

	// in native Substation applications configuration is handled by compiling Jsonnet and loading JSON into the application
	cfg := []config.Config{
		{
			Type: "ip_database",
			Settings: map[string]interface{}{
				"input_key":  "ip",
				"output_key": "as",
				"options": map[string]interface{}{
					"function": "maxmind_asn",
					"database_options": map[string]interface{}{
						"type": "maxmind_asn",
						"settings": map[string]interface{}{
							// the location of the IP enrichment database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
							"database": "location://path/to/maxmind.mmdb",
							"language": "en",
						},
					},
				},
			},
		},
		{
			Type: "ip_database",
			Settings: map[string]interface{}{
				"input_key":  "ip",
				"output_key": "geo",
				"options": map[string]interface{}{
					"function": "maxmind_city",
					"database_options": map[string]interface{}{
						"type": "maxmind_city",
						"settings": map[string]interface{}{
							// the location of the IP enrichment database can be either a path on local disk, an HTTP(S) URL, or an AWS S3 URL
							"database": "location://path/to/maxmind.mmdb",
							"language": "en",
						},
					},
				},
			},
		},
	}

	applicators, err := process.MakeApplicators(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applicator
	defer process.CloseApplicators(context.TODO(), applicators...)

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
