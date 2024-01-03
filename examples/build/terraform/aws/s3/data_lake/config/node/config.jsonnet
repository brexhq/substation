local sub = import '../../../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'f926941a-30b6-f858-6f4b-7a48d8808ab3-substation';

{
  concurrency: 1,
  // All data is buffered in memory, then written in JSON Lines format to S3.
  transforms: [
    sub.tf.send.aws.s3(
      settings={
        auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
        bucket_name: bucket,
        path: { prefix: 'original' },
      }
    ),
    sub.tf.object.insert(
      settings={ object: { target_key: 'transformed' }, value: true }
    ),
    sub.tf.send.aws.s3(
      settings={
        auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
        bucket_name: bucket,
        path: { prefix: 'transformed' },
      }
    ),
  ],
}
