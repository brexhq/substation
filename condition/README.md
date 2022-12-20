# condition

Contains interfaces and methods for evaluating data using success or failure criteria. Conditions combine inspectors (e.g. string equals "foo", string matches "^foo") and an operator (e.g., AND, OR) to verify the state of data before applying other functions. Each inspector defines its own data processing patterns, but there are a set of common patterns shared among most inspectors:
* evaluating unstructured data
* evaluating JSON objects

The package can be used like this ([more examples are also available](/examples/condition/)):

```go
package main

import (
	"fmt"

	"github.com/brexhq/substation/condition"
)

func main() {
	inspector := condition.strings{
		Key:        "foo",
		Function:   "equals",
		Expression: "bar",
	}

	data := []byte(`{"foo":"bar"}`)
	ok, err := inspector.Inspect(data)
	if err != nil {
		panic(err)
	}

	if ok {
		fmt.Println("data passed inspection")
	} else {
		fmt.Println("data failed inspection")
	}
}
```

Information for each inspector and operator is available in the [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/condition).
