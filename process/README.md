# process
Contains interfaces and methods for atomically processing data. Processors can be applied to bytes and channels of bytes; for JSON data, some processors are array-aware and will automatically process data within arrays.

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

In Substation applications, processors adhere to these rules:
- share a common configuration syntax
  - input: settings that control where input is located (e.g., input.key)
  - output: settings that control where output is placed (e.g., output.key)
  - options: settings that control runtime options for the processor
- applied via conditions (`condition`)
- operate on JSON data

## processors
| Processor                           | Description |
| ---                                 | --- | 
| [Base64](#base64)                   | Encodes and decodes bytes to and from base64 |
| [Capture](#capture)                 | Applies a capturing regular expression |
| [Case](#case)                       | Modifies the case of a string |
| [Concat](#concat)                   | Modifies the case of a string |
| [Convert](#convert)                 | Converts a value to a new type (e.g., string to integer) |
| [Copy](#copy)                       | Copies a value from one JSON key to another |
| [Count](#count)                     | Count data in a channel |
| [Delete](#delete)                   | Deletes a JSON key |
| [Domain](#domain)                   | Parses a fully qualified domai name into separate labels (e.g., top-level domain, subdomain) |
| [Drop](#drop)                       | Drops data from a channel |
| [DynamoDB](#dynamodb)               | Runs a query on a DynamoDB table and returns matched items |
| [Expand](#expand)                   | Expands JSON arrays into individual objects |
| [Flatten](#flatten)                 | Flattens an array of values, including deeply nested arrays |
| [Gzip](#gzip)                       | Compresses and decompresses bytes to and from Gzip |
| [Hash](#hash)                       | Calculates the hash of a value |
| [Insert](#insert)                   | Inserts a value into a JSON key |
| [Lambda](#lambda)                   | Synchronously invokes an AWS Lambda and returns the results |
| [Math](#math)                       | Performs mathematical operations (e.g., add three values, subtract two values) |
| [Replace](#replace)                 | Replaces characters within a string |
| [Time](#time)                       | Converts time values between formats |
| [Zip](#zip)                         | Concatenates arrays into tuples or JSON objects |

### base64
Processes data by encoding it to or decoding it from base64. This processor should be used for converting entire JSON objects. The processor supports these base64 alphabets:
- std: https://www.rfc-editor.org/rfc/rfc4648.html#section-4
- url: https://www.rfc-editor.org/rfc/rfc4648.html#section-5

The processor uses this Jsonnet configuration:
```
{
  // if the input is `eyJoZWxsbyI6IndvcmxkIn0=`, then the output is `{"hello":"world"}`
  type: 'base64',
  settings: {
    options: {
      direction: 'from',
      alphabet: 'std',  // defaults to std
    }
  },
}
```

### capture
Processes data by applying a capturing regular expression. This processor is array-aware and can output one or many values that are automatically stored as values or arrays of elements.

The processor uses this Jsonnet configuration:
```
{
  type: 'capture',
  settings: {
    // if the value is "bar", then this returns ["b","a","r"]
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    },
    options: {
      expression: '(.{1})'
      count: 3,
    }
  },
}
```

### case
Processes data by converting the case of a string. This processor is array-aware and supports these options:
- upper: converts to uppercase
- lower: converts to lowercase
- snake: converts to [snake case](https://en.wikipedia.org/wiki/Snake_case)

The processor uses this Jsonnet configuration:
```
{
  type: 'case',
  settings: {
    // if the value is "bar", then this returns "BAR"
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    },
    options: {
      case: 'upper',
    }
  },
}
```

### concat
Processes data by concatenating multiple values together with a separator. This processor is array-aware.

The processor uses this Jsonnet configuration:
```
{
  type: 'concat',
  settings: {
    // if the values are "baz" and "qux", then this returns "baz.qux"
    input: {
      keys: ['foo','bar'],
    },
    output: {
      key: 'processed',
    },
    options: {
      separator: '.',
    }
  },
}
```

### convert
Processes data by converting values between types (e.g., string to integer, integer to float). This processor is array-aware and supports these types:
- bool: boolean
- int: integer
- float: float
- uint: uinteger
- string: string

The processor uses this Jsonnet configuration:
```
{
  type: 'convert',
  settings: {
    // if the value is "100", then this returns 100
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    },
    options: {
      type: 'int',
    }
  },
}
```

### copy
Processes data by copying it. The processor supports these patterns:
- json
  - `{"hello":"world"} >>> {"hello":"world","goodbye":"world"}`
- from json
  - `{"hello":"world"} >>> world`
- to json
  - `world >>> {"hello":"world"}`

The processor uses this Jsonnet configuration:
```
{
  type: 'copy',
  settings: {
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    },
  },
}
```

### count
Processes data by counting data in a channel. The output of this processor is `{"count":N}`, where `N` is the number of bytes that were in the channel.

The processor uses this Jsonnet configuration:
```
{
  type: 'count',
}
```

### delete
Processes data by deleting JSON keys. Any keys nested under the provided key are deleted.

The processor uses this Jsonnet configuration:
```
{
  type: 'delete',
  settings: {
    // if "foo" is in the JSON object, then this processor deletes it
    input: {
      key: 'foo',
    },
  },
}
```

### domain
Processes data by parsing fully qualified domain names into separate labels. This processor is array-aware and supports these options:
- tld: top-level domain (e.g., com)
- domain: tld + one label (e.g., brex.com)
- subdomain: subdomain (e.g., www)

The processor uses this Jsonnet configuration:
```
{
  type: 'domain',
  settings: {
    // if the value is "www.brex.com", then this returns "brex.com'
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    },
    options: {
      function: 'domain',
    }
  },
}
```

### drop
Processes data by dropping it from a channel. 

The processor uses this Jsonnet configuration:
```
{
  type: 'drop',
}
```

### dynamodb
Processes data by querying DynamoDB and returning all matched items as an array of JSON objects. This processor is array-aware.

We recommend referring to the [documentation for querying DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html) when working with this processor. Note that DynamoDB is designed for single-digit millisecond latency, but latency can takes 10s of milliseconds which can have significant impact on total event latency. If Substation is running in AWS Lambda with Kinesis, then this latency can be mitigated by increasing the [parallelization factor of the Lambda](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).

The processor uses this Jsonnet configuration:
```
{
  type: 'dynamodb',
  settings: {
    // if the value is "bar", then this queries DynamoDB by using "bar" as the paritition key value for the attribute "pk" and returns the last indexed item from the table.
    input: {
      partition_key: 'foo',
    },
    output: {
      key: 'processed',
    },
    options: {
      table: 'foo-table',
      key_condition_expression: 'pk = :partitionkeyval',
      // multiple items can be returned by changing limit
      limit: 1,
      // the order of the returned items can be changed by excluding scan_index_forward or setting it to false.
      scan_index_forward: true,
    }
  },
}
```

### expand
Processes data by expanding data in JSON arrays into individual events. This processor can optionally retain keys outside of the JSON array and insert them into the new events.

The processor uses this Jsonnet configuration:
```
{
  type: 'expand',
  settings: {
    // if the original event is {"foo":[{"bar":"baz"}],"qux":"quux"}, then this expands to create the event {"bar":"baz","qux":"quux"}
    input: {
      key: 'foo',
    },
    options: {
      retain: ['qux'],
    }
  },
}
```

### flatten
Processes data by flattening JSON arrays. This processor can optionally deeply flatten arrays.

The processor uses this Jsonnet configuration:
```
{
  type: 'flatten',
  settings: {
    // if the value is [1,2,[3,4,[5,6]]], then this returns [1,2,3,4,5,6]
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    }
    options: {
      deep: true,
    }
  },
}
```

### gzip
Processes data by compressing it to or decompressing it from gzip. This processor should be used for converting entire JSON objects.

The processor uses this Jsonnet configuration:
```
{
  type: 'gzip',
  settings: {
    options: {
      direction: 'from',
    }
  },
}
```

### hash
Processes data by calculating its hash. This processor is array-aware and supports these algorithms:
- md5
- sha256

The processor uses this Jsonnet configuration:
```
{
  type: 'hash',
  settings: {
    // calculates sha256 hash of value in "foo"
    // use "@this" to calculate the hash of entire JSON object
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    }
    options: {
      algorithm: 'sha256',
    }
  },
}
```

### insert
Processes data by inserting a value into a JSON object. This processor supports any type of value.

The processor uses this Jsonnet configuration:
```
{
  type: 'insert',
  settings: {
    // inserts value "foo" into key "processed"
    output: {
      key: 'processed',
    }
    options: {
      value: 'foo',
    }
  },
}
```

### lambda
Processes data by synchronously invoking an AWS Lambda and returning the results as a JSON object. This processor optionally treats errors in the invoked Lambda as errors in the processor (by default, if errors occur then they are ignored and the input data is returned).

Note that the average latency of synchronously invoking a Lambda function is 10s of milliseconds, but latency can take 100s to 1000s of milliseconds depending on the function which can have significant impact on total event latency. If Substation is running in AWS Lambda with Kinesis, then this latency can be mitigated by increasing the [parallelization factor of the Lambda](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html).


The processor uses this Jsonnet configuration:
```
{
  type: 'lambda',
  settings: {
    // creates an AWS Lambda payload that maps keys from the input JSON object to keys in the payload
    input: {
      payload: [
        {
          key: 'foo',
          payload_key: 'ip_address',
        }
      ],
    },
    output: {
      key: 'processed',
    }
    options: {
      function: 'foo-function',
    }
  },
}
```

### math
Processes data by applying mathematical operations to multiple values. This processor supports these operations:
- add
- subtract

The processor uses this Jsonnet configuration:
```
{
  type: 'math',
  settings: {
    // if the values are 5 and 10, then this returns 15
    input: {
      keys: ['foo','bar'],
    },
    output: {
      key: 'processed',
    }
    options: {
      operation: 'add',
    }
  },
}
```

### replace
Processes data by replacing substrings in string values. This processor is array-aware.

The processor uses this Jsonnet configuration:
```
{
  type: 'replace',
  settings: {
    // if the value is "bar", then this returns "baz"
    // if the value is "barbar", then this returns "bazbar"
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    }
    options: {
      old: 'bar',
      new: 'baz',
      count: 1,  // defaults to 0, which replaces all substring matches
    }
  },
}
```

### time
Processes data by converting time values between formats. This processor is array-aware and supports these time formats:
- [pattern-based layouts](https://gobyexample.com/time-formatting-parsing)
- unix: epoch
- unix_milli: epoch milliseconds
- unix_nano: epoch nanoseconds
- now: current time

The processor uses this Jsonnet configuration:
```
{
  type: 'time',
  settings: {
    // if the value is 0, then this returns "1970-01-01T12:00:00"
    input: {
      key: 'foo',
    },
    output: {
      key: 'processed',
    }
    options: {
      input_format: 'epoch',
      output_format: '2006-01-02T15:04:05',
    }
  },
}
```

### zip
Processes data by concatenating JSON arrays into an array of tuples or array of JSON objects. 

For processing into an array of tuples, use this Jsonnet configuration:
```
{
  type: 'zip',
  settings: {
    // if the values are ["foo","bar"] and [123,456], then this returns [["foo",123],["bar",456]]
    input: {
      keys: ['names','sizes'],
    },
    output: {
      key: 'processed',
    }
  },
}
```

For processing into an array of JSON objects, use this Jsonnet configuration:
```
{
  type: 'zip',
  settings: {
    // if the values are ["foo","bar"] and [123,456], then this returns [{"name":"foo","size":123},{"name":"bar","size":456}]
    input: {
      keys: ['names','sizes'],
    },
    output: {
      key: 'processed',
    }
    options: {
      keys: ['name','size'],
    }
  },
}
```
