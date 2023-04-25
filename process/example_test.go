package process_test

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"
	"golang.org/x/sync/errgroup"
)

func ExampleNewApplier() {
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(applier)
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleNewAppliers() {
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more appliers are created
	appliers, err := process.NewAppliers(ctx, cfg)
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
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(ctx, cfg)
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
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more appliers are created
	batcher, err := process.NewBatcher(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(batcher)
	// Output: {"condition":{"operator":"","inspectors":null},"key":"foo","set_key":"bar","ignore_close":false,"ignore_errors":false}
}

func ExampleNewBatchers() {
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// one or more batchers are created
	batchers, err := process.NewBatchers(ctx, cfg)
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
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// batcher is retrieved from the factory
	batcher, err := process.NewBatcher(ctx, cfg)
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

func ExampleNewStreamer() {
	ctx := context.TODO()

	// copies the value of key "a" into key "z"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "a",
			"set_key": "z",
		},
	}

	// create a streamer from the config
	streamer, err := process.NewStreamer(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(streamer)
	// Output: {"condition":{"operator":"","inspectors":null},"key":"a","set_key":"z","ignore_close":false,"ignore_errors":false}
}

func ExampleNewStreamers() {
	ctx := context.TODO()

	// copies the value of key "a" into key "z"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "a",
			"set_key": "z",
		},
	}

	// one or more streamers are created
	streamers, err := process.NewStreamers(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	for _, s := range streamers {
		fmt.Println(s)
	}
	// Output: {"condition":{"operator":"","inspectors":null},"key":"a","set_key":"z","ignore_close":false,"ignore_errors":false}
}

func Example_applier() {
	ctx := context.TODO()

	// copies the value of key "foo" into key "baz"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// applier is retrieved from the factory
	applier, err := process.NewApplier(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// applier is applied to capsule
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"foo":"fizz"}`))

	capsule, err = applier.Apply(ctx, capsule)
	if err != nil {
		// handle err
		panic(err)
	}

	fmt.Println(string(capsule.Data()))
	// Output:
	// {"foo":"fizz","bar":"fizz"}
}

func Example_batcher() {
	ctx := context.TODO()

	// copies the value of key "foo" into key "bar"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "foo",
			"set_key": "bar",
		},
	}

	// batcher is retrieved from the factory
	batcher, err := process.NewBatcher(ctx, cfg)
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

func Example_streamer() {
	ctx := context.TODO()

	// copies the value of key "a" into key "z"
	cfg := config.Config{
		Type: "copy",
		Settings: map[string]interface{}{
			"key":     "a",
			"set_key": "z",
		},
	}

	// streamer is retrieved from the factory
	streamer, err := process.NewStreamer(ctx, cfg)
	if err != nil {
		// handle err
		panic(err)
	}

	// setup an error group and channels for streaming data
	group, ctx := errgroup.WithContext(ctx)
	in, out := config.NewChannel(), config.NewChannel()

	// first, the streamer goroutine is started
	group.Go(func() error {
		if err := streamer.Stream(ctx, in, out); err != nil {
			panic(err)
		}

		return nil
	})

	// second, the reader goroutine is started. reading must precede
	// writing to prevent blocking the main goroutine.
	group.Go(func() error {
		for capsule := range out.C {
			fmt.Println(string(capsule.Data()))
		}

		return nil
	})

	// last, writer sends data to the streamer.
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"a":"b"}`))

	in.Send(capsule)
	in.Close()

	// an error in any streamer will cause the entire pipeline, including sending
	// to the sink, to fail.
	if err := group.Wait(); err != nil {
		panic(err)
	}

	// Output:
	// {"a":"b","z":"b"}
}

func Example_dNS() {
	ctx := context.TODO()

	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"addr":"8.8.8.8"}`))

	// apply a reverse_lookup DNS query to addr
	cfg := []config.Config{
		{
			Type: "dns",
			Settings: map[string]interface{}{
				"key":     "addr",
				"set_key": "domains",
				"options": map[string]interface{}{
					"function": "reverse_lookup",
				},
			},
		},
	}

	appliers, err := process.NewAppliers(ctx, cfg...)
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

func Example_hTTP() {
	ctx := context.TODO()
	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"addr":"8.8.8.8"}`))

	// enriches the IP address by querying the GreyNoise Community API.
	// authenticating to GreyNoise is accomplished by interpolating a
	// secret inside an HTTP header. refer to the GreyNoise documentation
	// for more information:
	// https://docs.greynoise.io/reference/get_v3-community-ip.
	cfg := []config.Config{
		{
			Type: "http",
			Settings: map[string]interface{}{
				"key": "addr",
				// the HTTP response body is written to this key
				"set_key": "greynoise",
				"options": map[string]interface{}{
					"method": "get",
					// the value from "addr" is interpolated into ${data}
					"url": "https://api.greynoise.io/v3/community/${data}",
					"headers": []map[string]interface{}{
						{
							"key": "key",
							// this secret must be stored in the environment
							// variable GREYNOISE_API
							"value": "${SECRETS_ENV:GREYNOISE_API}",
						},
					},
				},
			},
		},
	}

	appliers, err := process.NewAppliers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applier
	defer process.CloseAppliers(ctx, appliers...)

	for _, app := range appliers {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	// sample output (which may change day to day)
	// {"addr":"8.8.8.8","greynoise":{"ip":"8.8.8.8","noise":false,"riot":true,"classification":"benign","name":"Google Public DNS","link":"https://viz.greynoise.io/riot/8.8.8.8","last_seen":"2023-01-30","message":"Success"}}
	fmt.Println(string(capsule.Data()))
}

func Example_iPDatabase() {
	ctx := context.TODO()
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

	appliers, err := process.NewAppliers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applier
	defer process.CloseAppliers(ctx, appliers...)

	for _, app := range appliers {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	fmt.Println(string(capsule.Data()))
}

func Example_kVStore() {
	ctx := context.TODO()

	capsule := config.NewCapsule()
	capsule.SetData([]byte(`{"foo":"bar"}`))

	// the value from key "foo" is first set into the KV store and
	// then retrieved from the KV store and set into key "baz". if
	// the KV options are identical across processors, then the same
	// KV store is used in each call. this also allows for the use of
	// multiple KV stores.
	cfg := []config.Config{
		{
			Type: "kv_store",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "setter",
				"options": map[string]interface{}{
					"type": "set",
					"kv_options": map[string]interface{}{
						"type": "memory",
						"settings": map[string]interface{}{
							"capacitiy": 10,
						},
					},
				},
			},
		},
		{
			Type: "kv_store",
			Settings: map[string]interface{}{
				"key":     "setter",
				"set_key": "baz",
				"options": map[string]interface{}{
					"type": "get",
					"kv_options": map[string]interface{}{
						"type": "memory",
						"settings": map[string]interface{}{
							"capacitiy": 10,
						},
					},
				},
			},
		},
	}

	appliers, err := process.NewAppliers(ctx, cfg...)
	if err != nil {
		// handle err
		panic(err)
	}

	//nolint: errcheck // errors are ignored in case processing fails in a single applier
	defer process.CloseAppliers(ctx, appliers...)

	for _, app := range appliers {
		capsule, err = app.Apply(ctx, capsule)
		if err != nil {
			// handle err
			panic(err)
		}
	}

	fmt.Println(string(capsule.Data()))
}
