local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.send.aws.sns(
      settings={topic: 'arn:aws:sns:us-east-1:123456789012:my-topic', aws: {region: 'us-east-1'}}
    ),
  ]
}
