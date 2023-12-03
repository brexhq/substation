local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.aws.sns(
      // This is a placeholder that must be replaced with the SNS ARN produced by Terraform.
      settings={ arn: 'arn:aws:sns:us-east-1:123456789012:substation' },
    ),
  ],
}
