# condition

Contains interfaces and methods for evaluating data for success or failure criteria. Conditions are a combination of operators (e.g., AND, OR) and inspectors (e.g. string equals "foo", regular expression matches "^foo") that can be used by applications that need to verify data before applying other processing functions. Each inspector defines its own data processing patterns, but there are a set of common patterns shared among most inspector:
- evaluating JSON values
- evaluating bytes

The package can be used like this:

```go
package main

import (
	"fmt"

	"github.com/brexhq/substation/condition"
)

func main() {
	inspector := condition.Strings{
		Key:        "hello",
		Expression: "world",
		Function:   "equals",
	}

	data := []byte(`{"hello":"world"}`)
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

Information for each operator and inspector can be found in the package's [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/condition).
