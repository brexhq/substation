// This config generates an error to engage the retry on failure feature.
// The pipeline will retry forever until the error is resolved. Change the
// transform to `sub.tf.send.stdout()` to resolve the error and print the logs
// from S3.
local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.util.err(settings={ message: 'simulating error to trigger retries' }),
    // sub.tf.send.stdout(),
  ],
}
