# process
Contains interfaces and methods for atomically processing data. Each processor defines its own data processing patterns, but there are a set of common patterns shared among most processors:
- processing JSON values
- processing JSON arrays
- processing bytes

The package can be used like this ([more examples are also available](/examples/process/)):
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
		OutputKey: "foo",
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

Information for each processor is available in the [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/process).
