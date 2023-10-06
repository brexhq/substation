local sub = import '../../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'd7c66938-6b96-21a3-7e59-62cb40e4627f-substation';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.aws.s3(
      settings={ bucket_name: bucket, file_path: { prefix: 'original' } }
    ),
    sub.tf.object.insert(
      settings={ object: { set_key: 'transformed' }, value: true }
    ),
    sub.tf.send.aws.s3(
      settings={ bucket_name: bucket, file_path: { prefix: 'transformed' } }
    ),
  ],
}
