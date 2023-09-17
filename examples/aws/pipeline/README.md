# pipeline

This example deploys a data pipeline to AWS that makes use of every Substation component.

The data pipeline is visualized below:

```mermaid
graph TD
    %% core infrastructure
    dynamodb_table(DynamoDB Table)
    gateway_kinesis(API Gateway)
    kinesis_raw(Kinesis Data Stream)
    kinesis_processed(Kinesis Data Stream)
    s3_source_bucket(S3 Bucket)
    s3_lake_bucket(S3 Bucket)
    s3_bucket(S3 Bucket)
    sns_topic(SNS Topic)
    sqs_queue(SQS Queue)

    %% Lambda data processing
    dynamodb_lambda[Lambda]
    kinesis_lambda[Lambda]
    enrichment_lambda[Lambda]
    s3_warehouse_sink_lambda[Lambda]
    async_source_lambda[Lambda]
    s3_source_lambda[Lambda]
    s3_lake_sink_lambda[Lambda]
    sns_source_lambda[Lambda]
    sqs_source_lambda[Lambda]

    %% ingest
    async_source_lambda ---|Push| kinesis_raw
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
    kinesis_processed ---|Pull| s3_warehouse_sink_lambda ---|Push| s3_bucket
    kinesis_processed ---|Pull| dynamodb_lambda ---|Push| dynamodb_table
```

## Deployment 

See this recipe: https://substation.readme.io/recipes/deploying-example-aws-pipeline
