local sub = import '../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'd7c66938-6b96-21a3-7e59-62cb40e4627f-substation';

{
  concurrency: 1,
  metrics: {
    type: 'aws_cloudwatch_embedded_metrics',
  },
  transforms: [
    sub.transform.send.aws.s3(
		settings={ bucket: bucket, file_path: { prefix: 'raw' } }
    ),
	sub.transform.object.insert(
		settings={ object: { set_key: 'transformed' }, value: true }
	),
	sub.transform.send.aws.s3(
		settings={ bucket: bucket, file_path: { prefix: 'processed' } }
	),
  ],
}
