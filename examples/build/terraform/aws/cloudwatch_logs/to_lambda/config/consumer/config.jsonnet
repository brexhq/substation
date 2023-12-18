local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // CloudWatch logs sent to Lambda are base64 encoded and gzip
    // compressed within the `awslogs.data` field of the event. 
    // These must be decoded and decompressed before other transforms are
    // applied. 
    sub.tf.obj.cp({obj: {key: 'awslogs.data'}}),
    sub.tf.fmt.from.base64(),
    sub.tf.fmt.from.gzip(),
    sub.tf.send.stdout(),
  ],
}
