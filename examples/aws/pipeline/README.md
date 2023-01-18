# pipeline

This example deploys a data pipeline to AWS that makes use of every Substation component.

The data pipeline is visualized below:

```mermaid
graph TD
    %% core infrastructure
    dynamodb_table(Metadata DynamoDB Table)
    gateway_kinesis(HTTPS Endpoint)
    gateway(HTTPS Endpoint)
    kinesis_raw(Raw Data Kinesis Stream)
    kinesis_processed(Processed Data Kinesis Stream)
    s3_source_bucket(S3 Data Storage)
    s3_lake_bucket(Data Lake S3 Storage)
    s3_warehouse_bucket(Data Warehouse S3 Storage)
    sns_topic(SNS Topic)
    sqs_queue(SQS Queue)

    %% Lambda data processing
    dynamodb_lambda[DynamoDB Sink Lambda]
    gateway_lambda[Gateway Source Lambda]
    kinesis_lambda[Processor Lambda]
    enrichment_lambda[Data Enrichment Lambda]
    s3_warehouse_sink_lambda[S3 Sink Lambda]
    async_source_lambda[Async Source Lambda]
    s3_source_lambda[S3 Source Lambda]
    s3_lake_sink_lambda[S3 Sink Lambda]
    sns_source_lambda[SNS Source Lambda]
    sqs_source_lambda[SQS Source Lambda]

    %% ingest
    async_source_lambda ---|Push| kinesis_raw
    gateway ---|Push| gateway_lambda ---|Push| kinesis_raw
    gateway_kinesis ---|Push| kinesis_raw
    s3_source_bucket ---|Pull| s3_source_lambda ---|Push| kinesis_raw
    sns_topic ---|Push| sns_source_lambda ---|Push| kinesis_raw
    sqs_queue ---|Pull| sqs_source_lambda ---|Push| kinesis_raw
    kinesis_raw ---|Pull| s3_lake_sink_lambda ---|Push| s3_lake_bucket

    %% transform
    kinesis_raw ---|Pull| kinesis_lambda ---|Push| kinesis_processed
    kinesis_lambda ---|Pull| dynamodb_table
    kinesis_lambda ---|Invoke| enrichment_lambda

    %% load
    kinesis_processed ---|Pull| s3_warehouse_sink_lambda ---|Push| s3_warehouse_bucket
    kinesis_processed ---|Pull| dynamodb_lambda ---|Push| dynamodb_table
```

## Deployment 

See this recipe: https://substation.readme.io/recipes/deploying-example-aws-pipeline
