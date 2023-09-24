local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  metrics: {
    type: 'aws_cloudwatch_embedded_metrics',
  },
  transforms: [
    sub.transform.send.aws.kinesis_data_stream(
      settings={ stream_name: 'dst' },
    ),
  ],
}
