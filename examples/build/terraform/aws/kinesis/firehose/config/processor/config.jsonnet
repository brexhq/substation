local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.object.insert(
      settings={ object: { dst: 'transformed' }, value: true }
    ),
    // Appending a newline is required so that the S3 object is line delimited.
    sub.tf.string.append(
      settings={ suffix: '\n' }
    ),
  ],
}
