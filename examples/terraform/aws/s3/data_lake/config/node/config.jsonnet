local sub = import '../../../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'c034c726-70bf-c397-81bd-c9a0d9e82371-substation';

{
  concurrency: 1,
  // All data is buffered in memory, then written in JSON Lines format to S3.
  transforms: [
    sub.tf.send.aws.s3(
      settings={
        bucket_name: bucket,
        file_path: { prefix: 'original', time_format: '2006/01/02', uuid: true },
        auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
      }
    ),
    sub.tf.object.insert(
      settings={ object: { target_key: 'transformed' }, value: true }
    ),
    sub.tf.send.aws.s3(
      settings={
        bucket_name: bucket,
        file_path: { prefix: 'transformed', time_format: '2006/01/02', uuid: true },
        auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
      }
    ),
  ],
}
