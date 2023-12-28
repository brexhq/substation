local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.pattern.transform.conditional(
      transform=sub.transform.object.copy(
        settings={ object: { src: 'body' } }
      ),
      condition=sub.condition.all([
        sub.condition.number.length.greater_than(
          settings={ object: { src: 'body' }, value: 0 }
        ),
      ]),
    ),
    // This UUID is used by the client to retrieve the processed result from DynamoDB.
    sub.transform.string.uuid(
      settings={ object: { dst: 'uuid' } },
    ),
    sub.transform.send.aws.sqs(
      // This is a placeholder that must be replaced with the SQS ARN produced by Terraform.
      settings={ arn: 'arn:aws:sqs:us-east-1:123456789012:substation' },
    ),
  ],
}
