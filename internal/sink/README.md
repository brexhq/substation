# sink
Contains interfaces and methods for sinking data to external services. As a general rule, sinks should support any data, not only JSON, when possible. Each sink must select from both the data and kill channels to prevent goroutine leaks (learn more about goroutine leaks [here](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)).

| Sink                        | Description |
| ---                         | --- |
| [DynamoDB](#dynamodb)      | sink data to DynamoDB tables |
| [HTTP](#http)              | sink data to an HTTP(S) endpoint |
| [Kinesis](#kinesis)        | sink KPL-compliant aggregated records to a Kinesis Data Stream | 
| [S3](#s3)                  | sink data as a gzip object to an S3 bucket |
| [Stdout](#stdout)          | sink data to stdout |
| [Sumo Logic](#sumologic)   | sink data to Sumo Logic as an aggregated HTTP payload |

## dynamodb
Sinks data to a DynamoDB table. This sink makes no assumptions about the structure of the table; if you are unfamiliar with DynamoDB and/or NoSQL design, then we recommend [reading this best practices guide from AWS](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html).

The sink uses this Jsonnet configuration (see `config/example/substation_example_dynamodb/` for more examples):
```
{
  type: 'dynamodb',
  settings: {
    // provides support for sinking data to multiple DynamoDB tables; the sink applies the last successfully matched attribute in the array
    items: [
      {
        // condition used to match incoming events; see condition/README.md for more information
        condition: {
          inspectors: [
            conditionlib.strings.empty('event.hash', negate=true),
          ],
        },
        // DynamoDB table that the data is written to
        table: 'substation_example',
        // event keys that map to DynamoDB attributes / columns; this translates to "store the value from event.hash into attribute pk, store the value from event.created into attribute event_created"
        // DynamoDB tables 
        fields: [
          {
            key: 'event.hash',
            attribute: 'pk',
          },
          {
            key: 'event.created',
            attribute: 'event_created',
          },
        ],
      },
    ],
  },
}
```

## http
Sinks data to an HTTP(S) endpoint. This sink optionally supports the ability to insert values from JSON events into HTTP headers.

The sink uses this Jsonnet configuration:
```
{
  type: 'http',
  settings: {
    url: 'example.com/foo',
    // event keys that map to HTTP headers; this translates to "use the value from content.type as the HTTP header Content-Type"
    // some logging platforms
    headers: [
      {
        key: 'content.type',
        header: 'Content-Type',
      },
    ],
  },
}
```

## kinesis
Sinks data to a Kinesis Data Stream as a Kinesis Producer Library (KPL) compliant aggregated record. This sink optionally supports the ability to pull a stream partition key from JSON events, but by default uses random partition keys to [avoid hot shards](https://aws.amazon.com/blogs/big-data/under-the-hood-scaling-your-kinesis-data-streams/). 

The sink uses this Jsonnet configuration (see `config/example/substation_example_kinesis/` for more examples):
```
{
  type: 'kinesis',
  settings: {
    stream: 'substation_processed_example',
  },
}
```

## s3
Sinks data to an S3 bucket as newline-delimited, gzip compressed objects organized by year, month, and day. This sink optionally supports a key prefix that can be utilized to store multiple datasets in a single S3 bucket.

The sink uses this Jsonnet configuration (see `config/example/substation_example_s3_sink/` for more examples):
```
{
  type: 's3',
  settings: {
    bucket: 'substation-example-sink',
    prefix: 'example',
  },
}
```

## stdout
Sinks data to stdout. This sink is intended to be used during development and testing. 

The sink uses this Jsonnet configuration:
```
{
  type: 'stdout',
}
```

## sumologic
Sinks data to a Sumo Logic HTTP source as an [aggregated HTTP payload](https://help.sumologic.com/03Send-Data/Sources/02Sources-for-Hosted-Collectors/HTTP-Source/Upload-Data-to-an-HTTP-Source#upload-log-data-with-a-post-request). 

This sink uses this Jsonnet configuration:
```
{
  type: 'sumologic',
  settings: {
    // url is defined in the Sumo Logic console
    url: 'https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/foo',
    // provides support for sinking data to multiple Sumo Logic source categories; the sink applies the last successfully matched category in the array
    categories: [
      // if no other matches succeed, then all events are sent to this source category
      {
        category: 'foo/default',
      },
      {
        category: 'foo/bar',
        condition: {
          inspectors: [
            conditionlib.strings.equals('some.key', 'bar'),
          ],
        },
      },
      {
        category: 'foo/baz',
        condition: {
          inspectors: [
            conditionlib.strings.equals('some.key', 'baz'),
          ],
        },
      },
      {
        category: 'foo/qux',
        condition: {
          inspectors: [
            conditionlib.strings.equals('some.key', 'qux'),
          ],
        },
      },
    ],
  },
}
```
