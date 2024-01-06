local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.pattern.transform.conditional(
      condition=sub.condition.all([
        sub.condition.number.length.greater_than(
          settings={ object: { source_key: 'body' }, value: 0 }
        ),
      ]),
      transform=sub.transform.object.copy(
        settings={ object: { source_key: 'body' } }
      ),
    ),
    // This UUID is used by the client to retrieve the processed result from DynamoDB.
    sub.transform.string.uuid(
      settings={ object: { target_key: 'uuid' } },
    ),
    sub.transform.send.aws.sqs(
      // This is a placeholder that must be replaced with the SQS ARN produced by Terraform.
      settings={ arn: 'arn:aws:sqs:us-east-1:123456789012:substation' },
    ),
  ],
}
