local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  metrics: {
    type: 'aws_cloudwatch_embedded_metrics',
  },
  transforms: [
    sub.transform.send.aws.sns(
      // This topic is a placeholder that must be replaced with the topic produced by Terraform.
      settings={topic: 'arn:aws:sns:us-east-1:123456789012:substation', aws: {region: 'us-east-1'}}
    ),
  ]
}
