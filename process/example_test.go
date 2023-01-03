package process_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
)

func ExampleNewApplier() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(applier)
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleNewAppliers() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more appliers are created
	appliers, err := process.NewAppliers(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, app := range appliers {
		fmt.Println(app)
	}
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleApplyBytes() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// applier is applied to bytes
	b := []byte(`{"foo":"fizz"}`)
	b, err = process.ApplyBytes(context.TODO(), b, applier)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(string(b))
	// Output: {"foo":"fizz","bar":"fizz"}
}

func ExampleNewBatcher() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more appliers are created
	batcher, err := process.NewBatcher(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(batcher)
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleNewBatchers() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more batchers are created
	batchers, err := process.NewBatchers(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, bat := range batchers {
		fmt.Println(bat)
	}
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleBatchBytes() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// batcher is retrieved from the factory
	batcher, err := process.NewBatcher(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// applier is applied to slice of bytes
	b := [][]byte{[]byte(`{"foo":"fizz"}`)}
	b, err = process.BatchBytes(context.TODO(), b, batcher)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, e := range b {
		fmt.Println(string(e))
	}
	// Output: {"foo":"fizz","bar":"fizz"}
}

func Example_applier() {
	// copies the value of key "foo" into key "baz"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// applier is applied to capsule
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"foo":"fizz"}`))

	capsule, err = applier.Apply(context.TODO(), capsule)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(string(capsule.Data()))
	// Output:
	// {"foo":"fizz","bar":"fizz"}
}

func Example_batcher() {
	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// batcher is retrieved from the factory
	batcher, err := process.NewBatcher(cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// batcher is applied to capsules
	var capsules []config.Capsule
	capsule := config.NewCapsule()

	// {"foo":"fizz","bar":"buzz"}
	for i := 1; i < 3; i++ {
		_ = capsule.Set("foo", "fizz")
		_ = capsule.Set("bar", "buzz")
		capsules = append(capsules, capsule)
	}

	capsules, err = batcher.Batch(context.TODO(), capsules...)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, c := range capsules {
		fmt.Println(string(c.Data()))
	}

	// Output:
	// {"foo":"fizz","bar":"fizz"}
	// {"foo":"fizz","bar":"fizz"}
}

func Example_dNS() {
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"addr":"8.8.8.8"}`))

	// apply a reverse_lookup DNS query to addr
	cfg := []config.Config{
		{
			Type: "dns",
			Settings: map[string]interface{}{
				"input_key":  "addr",
				"output_key": "domains",
				"options": map[string]interface{}{
					"function": "reverse_lookup",
				},
			},
		},
	}

	appliers, err := process.NewAppliers(cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applier
	defer process.CloseAppliers(context.TODO(), appliers...)

	for _, app := range appliers {
		capsule, err = app.Apply(context.TODO(), capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	fmt.Println(string(capsule.Data()))
}

func Example_iPDatabase() {
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"addr":"8.8.8.8"}`))

	// lookup addr in MaxMind City database
	cfg := []config.Config{
		{
			Type: "ip_database",
			Settings: map[string]interface{}{
				"key":     "addr",
				"set_key": "geo",
				"options": map[string]interface{}{
					"type": "maxmind_city",
					"settings": map[string]interface{}{
						// the location of the IP enrichment database can
						// be either a path on local disk, an HTTP(S) URL,
						// or an AWS S3 URL
						"database": "location://path/to/maxmind.mmdb",
						"language": "en",
					},
				},
			},
		},
	}

	appliers, err := process.NewAppliers(cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applier
	defer process.CloseAppliers(context.TODO(), appliers...)

	for _, app := range appliers {
		capsule, err = app.Apply(context.TODO(), capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	fmt.Println(string(capsule.Data()))
}
