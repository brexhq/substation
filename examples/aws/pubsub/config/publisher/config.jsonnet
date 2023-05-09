local sub = import '../../../../../build/config/substation.libsonnet';

{
  sink: sub.interfaces.sink.aws_sns(
    // change SNS topic ARN to match the resource created by Terraform
    settings = { arn: 'arn:aws:sns:us-east-1:123456789012:my-topic'}
  ),
  transform: {
    type: 'noop',
  },
}
