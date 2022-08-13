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
	proc := process.Insert{
		OutputKey: "baz",
		Options: process.InsertOptions{
			Value: "qux",
		},
	}

	data := []byte(`{"foo":"bar"}`)
	data, err := process.Byte(context.TODO(), proc, data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", data)
}
```

Information for each processor is available in the [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/process).
