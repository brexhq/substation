################################################
# KMS read permissions
# all Lambda must have this policy
################################################

module "iam_kms_read" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
  ]
}

module "iam_kms_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_read"
  policy = module.iam_kms_read.kms_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_enrichment.role,
    module.lambda_processor.role,
    module.lambda_dynamodb_sink.role,
    module.lambda_processed_s3_sink.role,
    module.lambda_raw_s3_sink.role,
    module.lambda_async_source.role,
    module.lambda_gateway_source.role,
    module.lambda_s3_source.role,
    module.lambda_sns_source.role,
    module.lambda_sqs_source.role,
    module.gateway_kinesis_source.role,
  ]
}

################################################
# KMS write permissions
# all Lambda must have this policy
################################################

module "iam_kms_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
  ]
}

module "iam_kms_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_write"
  policy = module.iam_kms_write.kms_write_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_enrichment.role,
    module.lambda_processor.role,
    module.lambda_dynamodb_sink.role,
    module.lambda_processed_s3_sink.role,
    module.lambda_raw_s3_sink.role,
    module.lambda_async_source.role,
    module.lambda_gateway_source.role,
    module.lambda_s3_source.role,
    module.lambda_sns_source.role,
    module.lambda_sqs_source.role,
    module.gateway_kinesis_source.role,
  ]
}
