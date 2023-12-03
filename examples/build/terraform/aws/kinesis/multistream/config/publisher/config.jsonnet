local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.aws.kinesis_data_stream(
      settings={ stream_name: 'dst' },
    ),
  ],
}
