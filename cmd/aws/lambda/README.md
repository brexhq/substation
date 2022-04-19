# lambda
Contains Substation apps deployed as AWS Lambda functions. All Lambda functions get their configurations from [AWS AppConfig](https://docs.aws.amazon.com/appconfig/latest/userguide/what-is-appconfig.html).

## autoscaling
This app handles Kinesis Data Stream autoscaling through SNS notifications and CloudWatch alarms. By default, Kinesis Data Streams will scale up by 50% if a stream exceeds 50% capacity within 1 minute and scale down by 50% if a stream stays below 15% capacity for 1 hour. Capacity is calculated as the maximum of volume (i.e., 60,000 events per minute) or size (i.e., 10GB data per minute).

For example, if a stream is configured with 10 shards and it exceeds 50% capacity, then the stream is scaled up to 15 shards; if the stream stays below 15% capacity for 1 hour, then the stream is scaled down to 5 shards. Shards will not scale evenly, but the autoscaling functionality follows [AWS best practices for resharding streams](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_UpdateShardCount.html).

## substation
This app handles ITL for data from these AWS services:
- API Gateway (REST)
- Kinesis Data Streams
- S3 Buckets
- S3 Buckets via SNS Topics
