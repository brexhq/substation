# condition

Contains interfaces and methods for evaluating data for user-defined success or failure criteria. Conditions are a combination of operators (e.g., AND, OR) and inspectors (e.g. string equals "foo", regular expression matches "^foo") that can be used by applications that need to verify data before applying other processing functions.

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

In Substation applications, conditions adhere to these rules:

- condition must pass (return true) to apply processors
- NOT statements are achieved by using negation operators (NAND, NOR) or negation settings in inspectors

## operators

Conditions rely on [boolean operators](https://en.wikipedia.org/wiki/Boolean_expression) to evaluate data. When using the OperatorFactory, the default behavior is to always return true no matter what data is provided as input.

| Operator | Description                              |
| -------- | ---------------------------------------- |
| AND      | All inspectors must return true to pass  |
| OR       | Any inspector must return true to pass   |
| NAND     | All inspectors must return false to pass |
| NOR      | Any inspector must return false to pass  |

## inspectors

Conditions use inspectors, which are atomic data inspection methods, to evaluate data. Inspectors can be independently used in other non-Substation applications to evaluate data. By default, all inspectors return false if the evaluation does not succeed, but this can be flipped by setting `negate` to true.

| Inspector                  | Description                               |
| -------------------------- | ----------------------------------------- |
| [Content](#content)        | Evaluates data by content type             |
| [IP](#ip)                  | Evaluates an IP address by type and usage |
| [JSONSchema](#json_schema) | Evaluates JSON key values by type         |
| [JSONValid](#json_valid)   | Evaluates whether data is valid JSON      |
| [RegExp](#regexp)          | Evaluates data with a regular expression  |
| [Strings](#strings)        | Evaluates data with string functions      |

### content

Inspects bytes and evalutes them by content type. This inspector uses the standard library's `net/http` package to identify the content type of data (more information is available [here](https://pkg.go.dev/net/http#DetectContentType)) and is most effective when using processors that change the format of data (e.g., `process/gzip`). The inspector supports these evaluations:

- type: the MIME type of the content (e.g., application/zip, application/x-gzip)

The inspector uses this Jsonnet configuration:

```
// returns true if the IP address value stored in JSON key "foo" is not a private address
{
  type: 'content',
  settings: {
    type: 'application/zip',
    negate: true,
  },
}
```

### ip

Inspects IP addresses and evaluates their type and usage. This inspector uses the standard library's `net` package to identify the type and usage of the address (more information is available [here](https://pkg.go.dev/net#IP)). The inspector supports these evaluations:

- loopback: valid loopback address
- multicast: valid multicast address
- multicast_link_local: valid link local multicast address
- private: valid private address
- unicast_global: valid global unicast address
- unicast_link_local: valid link local unicast address
- unspecified: valid "unspecified" address (e.g., 0.0.0.0, ::)

The inspector uses this Jsonnet configuration:

```
// returns true if the IP address value stored in JSON key "foo" is not a private address
{
  type: 'ip',
  settings: {
    key: 'foo',
    function: 'private',
    negate: true,
  },
}
```

### json_schema

Inspects JSON data and compares the key-value pairs against a user-defined schema. The inspector supports these schema types:

- string
- number (float, integer)
- boolean (true, false)
- json

The inspector uses this Jsonnet configuration:

```
// returns true if the value stored in JSON key "foo" is a string and the value stored in JSON key "bar" is an integer or float
{
  type: 'json_schema',
  settings: {
    negate: false,
    schema: [
      {
        key: "foo",
        type: "string",
      },
      {
        key: "bar",
        type: "number",
      }
    ],
  },
}
```

### json_valid

Inspects JSON data for validity. This inspector is a useful evaluation to use before attempting to sink data to systems that only support JSON.

The inspector uses this Jsonnet configuration:

```
// returns true if the data is valid JSON
{
  type: 'json_valid',
  settings: {
    negate: false,
  },
}
```

### regexp

Inspects data and evaluates it using a regular expression. This inspector uses a regexp cache provided by `internal/regexp`.

The inspector uses this Jsonnet configuration:

```
// returns true if the value stored in JSON key "foo" matches the regular expression
{
  type: 'regexp',
  settings: {
    key: 'foo',
    expression: '^bar',
    negate: false,
  },
}
```

### strings

Inspects data and evaluates it using string functions. This inspector uses the standard library's `strings` package. The inspector supports these string functions:

- equals: data equals the string expression
- contains: data contains the string expression
- endswith: data ends with the string expression
- startswith: data starts with the string expression

The inspector uses this Jsonnet configuration:

```
// returns true if the value stored in JSON key "foo" ends with the substring "bar"
{
  type: 'strings',
  settings: {
    key: 'foo',
    expression: 'bar',
    function: 'endswith',
    negate: false,
  },
}
```
