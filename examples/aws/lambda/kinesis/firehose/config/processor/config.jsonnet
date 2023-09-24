local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  metrics: {
    type: 'aws_cloudwatch_embedded_metrics',
  },
  transforms: [
    sub.transform.object.insert(
      settings={ object: { set_key: 'transformed' }, value: true }
    ),
    // Appending a newline is required so that the S3 object is line delimited.
    sub.transform.string.append(
      settings={ string: '\n' }
    ),
  ],
}