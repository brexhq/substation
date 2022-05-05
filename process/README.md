# process
Contains interfaces and methods for atomically processing data. Each processor defines its own data processing patterns, but there are a set of common patterns shared among most processors:
- processing JSON values, including arrays
- processing bytes

The package can be used like this:
```go
package main

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/process"
)

func main() {
	processor := process.Insert{
		Options: process.InsertOptions{
			Value: "bar",
		},
		Output: process.Output{
			Key: "foo",
		},
	}

	ctx := context.TODO()
	data := []byte(`{"hello":"world"}`)
	processed, err := processor.Byte(ctx, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	fmt.Println(string(processed))
}
```

Information for each processor can be found in the package's [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/process).
