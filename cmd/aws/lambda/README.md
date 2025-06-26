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
* [S3 via SQS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ways-to-add-notification-config-to-bucket.html)
* [SNS](https://docs.aws.amazon.com/lambda/latest/dg/with-sns.html)
* [SQS](https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html)

## autoscale

This app handles Kinesis Data Stream autoscaling through SNS notifications and CloudWatch alarms. Scaling is based on stream capacity as determined by the number and size of incoming records written to the stream. By default, the scaling behavior follows this pattern:

* If stream utilization is greater than 70% of the Kinesis service limits consistently within a 5 minute period, then scale up
* If stream utilization is less than 35% of the Kinesis service limits consistently within a 60 minute period, then scale down

The scaling behavior is customizable using environment variables:

* `AUTOSCALE_KINESIS_THRESHOLD` - The target threshold to cause a scaling event. The default value is 0.7 (70%), but it can be set to any value between 0.4 (40%) and 0.9 (90%). If the threshold is low, then the stream is more sensitive to scaling up and less sensitive to scaling down. If the threshold is high, then the stream is less sensitive to scaling up and more sensitive to scaling down.
* `AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS` - The number of data points required to scale up. The default value is 5, but it can be set to any value between 1 and 30. The number of data points affects the evaluation period; every 5 data points is equivalent to 5 minutes and the maximum evaluation period is 30 minutes. Use a higher value to reduce the frequency of scaling up.
* `AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS` - The number of data points required to scale down. The default value is 60, but it can be set to any value between 1 and 360. The number of data points affects the evaluation period; every 60 data points is equivalent to 1 hour and the maximum evaluation period is 6 hours. Use a higher value to reduce the frequency of scaling down.

Shards do not scale evenly, but the autoscaling follows [AWS best practices for resharding streams](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_UpdateShardCount.html). UpdateShardCount has many limitations that the application is designed around, but there may be times when these limits cannot be avoided; if any limits are met, then users should file a service limit increase with AWS. Although rare, the most common service limits that users may experience are:

* Scaling a stream more than 10 times per 24 hour rolling period
* Scaling a stream beyond 10000 shards

We recommend using one autoscaling Lambda for an entire Substation deployment, but many can be used if needed. For example, one can be assigned to data pipelines that have predictable traffic (e.g., steady stream utilization) and another can be assigned to data pipelines that have unpredictable traffic (e.g., sporadic stream utilization, bursty stream utilization).

## validate

This app handles checking if a configuration for the Substation app is valid without processing any data. It supports input from these methods:

* [AppConfig Validator Lambda](https://docs.aws.amazon.com/appconfig/2019-10-09/APIReference/API_Validator.html)
* [Lambda Invocation](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html)
