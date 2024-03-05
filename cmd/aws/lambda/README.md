# lambda

Contains Substation apps deployed as AWS Lambda functions. All Lambda functions get their configurations from [AWS AppConfig](https://docs.aws.amazon.com/appconfig/latest/userguide/what-is-appconfig.html) or AWS S3.

## substation

This app handles ingest, transform, and load for data from these AWS services:
* [API Gateway](https://docs.aws.amazon.com/lambda/latest/dg/services-apigateway.html)
* [DynamoDB Streams](https://docs.aws.amazon.com/lambda/latest/dg/with-ddb.html)
* [Kinesis Data Firehose](https://docs.aws.amazon.com/lambda/latest/dg/services-kinesisfirehose.html)
* [Kinesis Data Streams](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html)
* [Asynchronous Invocation (Lambda)](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html)
* [Synchronous Invocation (Lambda)](https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html)
* [S3](https://docs.aws.amazon.com/lambda/latest/dg/with-s3.html)
* [S3 via SNS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ways-to-add-notification-config-to-bucket.html)
* [SNS](https://docs.aws.amazon.com/lambda/latest/dg/with-sns.html)
* [SQS](https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html)

## autoscale

This app handles Kinesis Data Stream autoscaling through SNS notifications and CloudWatch alarms. The scaling behavior is to scale up / out if stream utilization is greater than 75% of the Kinesis service limits within a 5 minute period and scale down / in if stream utilization is less than 25% of the Kinesis service limits within a 60 minute period. In both cases, streams scale by 50%.

Stream utilization is based on volume (i.e., 60, 000 events per minute) and size (i.e., 10GB data per minute); these values are converted to a percentage (0.0 to 1.0) and the maximum of either is considered the stream's current utilization.

By default, streams must be above the upper threshold for all 5 minutes to scale up and below the lower threshold for at least 57 minutes to scale down. These values can be overriden by the environment variables AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS (cannot exceed 5 minutes) and AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS (cannot exceed 60 minutes).

For example:

* If a stream is configured with 10 shards and it triggers the upscale alarm, then the stream is scaled up to 15 shards
* If a stream is configured with 10 shards and it triggers the downscale alarm, then the stream is scaled down to 5 shards

Shards will not scale evenly, but the autoscaling functionality follows [AWS best practices for resharding streams](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_UpdateShardCount.html). UpdateShardCount has many limitations that the application is designed around, but there may be times when these limits cannot be avoided; if any limits are met, then users should file a service limit increase with AWS. Although rare, the most common service limits that users may experience are:

* Scaling a stream more than 10 times per 24 hour rolling period
* Scaling a stream beyond 10000 shards

We recommend using one autoscaling Lambda for an entire Substation deployment, but many can be used if needed. For example, one can be assigned to data pipelines that have predictable traffic (e.g., steady stream utilization) and another can be assigned to data pipelines that have unpredictable traffic (e.g., sporadic stream utilization, bursty stream utilization).

## validate

This app handles checking if a configuration for the Substation app is valid without processing any data. It supports input from these methods:

* [AppConfig Validator Lambda](https://docs.aws.amazon.com/appconfig/2019-10-09/APIReference/API_Validator.html)
* [Lambda Invocation](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html)
