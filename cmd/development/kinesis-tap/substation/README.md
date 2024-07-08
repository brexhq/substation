# kinesis-tap/substation

`kinesis-tap` is a tool for tapping into and transforming data from an AWS Kinesis Data Stream in real-time with Substation.

This is intended as a Substation development aid, but it has other uses as well, such as:
- Previewing live data in a stream by printing it to the console (default behavior)
- Sampling live data from a stream and saving it to a local file
- Forwarding live data between data pipeline stages for testing

Warning: This is a development tool intended to provide temporary access to live data in a Kinesis Data Stream; if you need to process data from a Kinesis Data Stream with strict reliability guarantees, use the [AWS Lambda applications](/cmd/aws/lambda/).

## Usage

```
% ./substation -h
Usage of ./substation:
  -config string
        The Substation configuration file used to transform records (default "./config.json")
  -stream-name string
        The AWS Kinesis Data Stream to fetch records from
  -stream-offset string
        Determines the offset of the stream (earliest, latest) (default "earliest")
```

Use the `SUBSTATION_DEBUG=1` environment variable to enable debug logging:
```
% SUBSTATION_DEBUG=1 ./substation -stream-name my-stream
DEBU[0000] Retrieved active shards from Kinesis stream.  count=2 stream=my-stream
DEBU[0001] Retrieved records from Kinesis shard.         count=981 shard=0x140004a6f80 stream=my-stream
DEBU[0002] Retrieved records from Kinesis shard.         count=1055 shard=0x140004a6fe0 stream=my-stream
DEBU[0003] Retrieved records from Kinesis shard.         count=2333 shard=0x140004a6f80 stream=my-stream
DEBU[0003] Retrieved records from Kinesis shard.         count=1110 shard=0x140004a6fe0 stream=my-stream
DEBU[0004] Retrieved records from Kinesis shard.         count=2109 shard=0x140004a6f80 stream=my-stream
DEBU[0004] Retrieved records from Kinesis shard.         count=1094 shard=0x140004a6fe0 stream=my-stream
^CDEBU[0004] Closed connections to the Kinesis stream.    
DEBU[0004] Closed Substation pipeline.                  
DEBU[0004] Flushed Substation pipeline.
```

## Build

```
git clone github.com/brexhq/substation && \
cd substation/cmd/development/kinesis-tap/substation && \
go build .
```

## Authentication

`kinesis-tap` uses the AWS SDK for Go to authenticate with AWS. The SDK uses the same authentication methods as the AWS CLI, so you can use the same environment variables or configuration files to authenticate.

For more information, see the [AWS CLI documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).
