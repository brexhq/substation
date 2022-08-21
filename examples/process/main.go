// example from process/README.md
package main

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/process"
)

func main() {
	proc := process.Insert{
		OutputKey: "baz",
		Options: process.InsertOptions{
			Value: "qux",
		},
	}

	data := []byte(`{"foo":"bar"}`)
	data, err := process.ApplyByte(context.TODO(), data, proc)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", data)
}
