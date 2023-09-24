local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.patterns.transform.conditional(
      transform=sub.transform.object.copy(
        settings={ object: { key: 'body' } }
      ),
      condition=sub.condition.all([
        sub.condition.logic.len.greater_than(
          settings={ object: { key: 'body' }, length: 0 }
        ),
      ]),
    ),
    // This UUID is used by the client to retrieve the processed result from DynamoDB.
    sub.transform.string.uuid(
      settings={ object: { set_key: 'uuid' } },
    ),
    sub.transform.send.aws.sqs(
      // This is a placeholder that must be replaced with the SQS ARN produced by Terraform.
      settings={ arn: 'arn:aws:sqs:us-east-1:123456789012:substation' },
    ),
  ],
}
