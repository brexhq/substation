package config_test

import (
	"fmt"
	"os"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
)

func ExampleGet_file() {
	// cfg is the location of a file on-disk
	cfg := config.Get()

	f, err := os.Open(cfg)
	if err != nil {
		// handle err
		panic(err)
	}
	defer f.Close()

	sub := cmd.New()
	if err := sub.SetConfig(f); err != nil {
		// handle err
		panic(err)
	}
}

func ExampleCapsule_Set() {
	data := []byte(`{"foo":"bar"}`)

	cap := config.NewCapsule()
	cap.SetData(data)

	cap.Set("baz", "qux")

	d := cap.Data()
	fmt.Println(string(d))
	// Output: {"foo":"bar","baz":"qux"}
}

func ExampleCapsule_SetData() {
	data := []byte(`{"foo":"bar"}`)

	cap := config.NewCapsule()
	cap.SetData(data)
}

func ExampleCapsule_SetMetadata() {
	metadata := struct {
		baz string
	}{
		baz: "qux",
	}

	cap := config.NewCapsule()
	cap.SetMetadata(metadata)
}

func ExampleCapsule_SetMetadata_chaining() {
	metadata := struct {
		baz string
	}{
		baz: "qux",
	}

	data := []byte(`{"foo":"bar"}`)

	cap := config.NewCapsule()
	cap.SetData(data).SetMetadata(metadata)
}
