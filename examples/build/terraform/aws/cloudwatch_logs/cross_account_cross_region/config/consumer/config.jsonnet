local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // CloudWatch logs sent to Kinesis Data Streams are gzip compressed. 
    // These must be decompressed before other transforms are applied.
    sub.tf.fmt.from.gzip(),
    sub.tf.send.stdout(),
  ],
}
