# transform

Contains interfaces and methods for transforming data between a source and a sink. Substation is designed for processing JSON data, but with processors (see `process/`) it can support any data format. Each transform must select from both the input and kill channels to prevent goroutine leaks (learn more about goroutine leaks [here](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)).

| Transform             | Description                           |
| --------------------- | ------------------------------------- |
| [Process](#process)   | applies processors to the data        |
| [Transfer](#transfer) | applies no transformation to the data |

## process

Transforms data by applying processors (`process/`). Processors are enabled by matching conditions (`condition/`) for every event. This transform iteratively processes individual events in the channel and modifies them before passing the event to the next processor.

Below is an example that shows how a single event is iteratively modified through this transform:

```
// input event
event: {"hello":"world"}

// insert value "bar" into key "foo"
processor: insert("foo", "bar")

event: {"hello":"world","foo":"bar"}

// insert value "qux" into key "foo"
processor: insert("baz", "qux")

event: {"hello":"world", "foo":"bar", "baz":"qux"}

// concat vaues from "foo" and "baz" into key "foo" with separator "."
processor: concat("foo", ["foo", "baz"], ".")

event: {"hello":"world", "foo":"bar.qux"}
```

The transform uses this Jsonnet configuration (see `config/example/substation_example_kinesis/` for more examples):

```
{
  type: 'process',
  settings: {
    // processors are defined according to the information in `process/`
    processors: [
      // adds an "event.created" key to the JSON event that contains the time that the processor executed
      {
        settings: {
          options: {
            input_format: 'now',
            output_format: '2006-01-02T15:04:05.000000Z',
          },
          output: {
            key: 'event.created',
          }
        },
        type: 'time',
      },
    ],
  },
}
```

## transfer

Transforms data with no modification, which we refer to as a "transfer." This transform is best used when the integrity of the original data needs to be maintained or if no data processing is needed.

The transform uses this Jsonnet configuration (see `config/example/` for more examples):

```
{
  type: 'transfer',
}
```
