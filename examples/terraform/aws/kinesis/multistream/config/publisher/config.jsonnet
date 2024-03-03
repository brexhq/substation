local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // This forwards data to the destination stream without transformation.
    sub.tf.send.aws.kinesis_data_stream(
      settings={ stream_name: 'substation_dst' },
    ),
  ],
}
