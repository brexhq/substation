################################################
# KMS read permissions
# all Lambda must have this policy
################################################

module "iam_example_kms_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
    aws_kms_key.xray_key.arn,
  ]
}

module "iam_example_kms_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kms_read"
  policy = module.iam_example_kms_read.kms_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_example_processor.role,
    module.lambda_example_dynamodb_sink.role,
    module.lambda_example_processed_s3_sink.role,
    module.lambda_example_raw_s3_sink.role,
    module.lambda_example_gateway_source.role,
    module.lambda_example_s3_source.role,
    module.gateway_example_kinesis_source.role,
  ]
}

################################################
# KMS write permissions
# all Lambda must have this policy
################################################

module "iam_example_kms_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
    aws_kms_key.xray_key.arn,
  ]
}

module "iam_example_kms_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kms_write"
  policy = module.iam_example_kms_write.kms_write_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_example_processor.role,
    module.lambda_example_dynamodb_sink.role,
    module.lambda_example_processed_s3_sink.role,
    module.lambda_example_raw_s3_sink.role,
    module.lambda_example_gateway_source.role,
    module.lambda_example_s3_source.role,
    module.gateway_example_kinesis_source.role,
  ]
}
