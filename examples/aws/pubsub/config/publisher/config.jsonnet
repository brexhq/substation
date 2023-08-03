local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.interfaces.transform.send.aws_sns(
      settings={topic: 'arn:aws:sns:us-east-1:123456789012:my-topic'}
    )    
  ]
}
