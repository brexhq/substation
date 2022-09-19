# lambda
Contains Substation apps deployed as AWS Lambda functions. All Lambda functions get their configurations from [AWS AppConfig](https://docs.aws.amazon.com/appconfig/latest/userguide/what-is-appconfig.html).

## autoscaling
This app handles Kinesis Data Stream autoscaling through SNS notifications and CloudWatch alarms. The scaling behavior is to scale up / out if stream utilization is greater than 75% of the Kinesis service limits within a 10 minute period and scale down / in if stream utilization are less than 25% of the Kinesis service limits within a 60 minute period. In both cases, streams scale by 50%.

Stream utilization is based on volume (i.e., 60,000 events per minute) and size (i.e., 10GB data per minute); these values are converted to a percentage (0.0 to 1.0) and the maximum of either is considered the stream's current utilization.

By default, streams must be above the upper threshold for at least 5 minutes to scale up and below the lower threshold for at least 57 minutes to scale down. These values can be overriden by the environment variables SUBSTATION_AUTOSCALING_UPSCALE_DATAPOINTS (cannot exceed 10 minutes) and SUBSTATION_AUTOSCALING_DOWNSCALE_DATAPOINTS (cannot exceed 60 minutes).

For example:

* If a stream is configured with 10 shards and it triggers the upscale alarm, then the stream is scaled up to 15 shards
* If a stream is configured with 10 shards and it triggers the downscale alarm, then the stream is scaled down to 5 shards

Shards will not scale evenly, but the autoscaling functionality follows [AWS best practices for resharding streams](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_UpdateShardCount.html). 

We recommend using one autoscaling Lambda for an entire Substation deployment, but many can be used if needed. For example, one may be assigned to data pipelines that have predictable traffic (e.g., steady stream utilization) and another may be assigned to data pipelines that have unpredictable traffic (e.g., sporadic stream utilization, bursty stream utilization).

## substation
This app handles ingest, transform, and load (ITL) for data from these AWS services:
- [API Gateway](https://docs.aws.amazon.com/lambda/latest/dg/services-apigateway.html)
- [Kinesis Data Streams](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html)
- [S3](https://docs.aws.amazon.com/lambda/latest/dg/with-s3.html)
- [S3 via SNS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ways-to-add-notification-config-to-bucket.html)
- [SNS](https://docs.aws.amazon.com/lambda/latest/dg/with-sns.html)
- [SQS](https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html)
