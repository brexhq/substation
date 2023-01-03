################################################
# raw data stream
################################################

module "kinesis_raw" {
  source            = "/workspaces/substation/build/terraform/aws/kinesis"
  kms_key_id        = module.kms_substation.arn
  stream_name       = "substation_raw"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn

  tags = {
    owner = "example"
  }
}

################################################
## raw data stream read permissions
################################################

module "iam_kinesis_raw_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_raw.arn,
  ]
}

module "iam_kinesis_raw_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_kinesis_raw_read"
  policy = module.iam_kinesis_raw_read.kinesis_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_processor.role,
    module.lambda_raw_s3_sink.role,
  ]
}

################################################
## raw data stream write permissions
################################################

module "iam_kinesis_raw_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_raw.arn,
  ]
}

module "iam_kinesis_raw_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_kinesis_raw_write"
  policy = module.iam_kinesis_raw_write.kinesis_write_policy
  roles = [
    module.lambda_gateway_source.role,
    module.lambda_s3_source.role,
    module.lambda_sns_source.role,
    module.lambda_sqs_source.role,
    module.gateway_kinesis_source.role,
  ]
}

################################################
# processed data stream
################################################

module "kinesis_processed" {
  source            = "/workspaces/substation/build/terraform/aws/kinesis"
  kms_key_id        = module.kms_substation.arn
  stream_name       = "substation_processed"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn

  tags = {
    owner = "example"
  }
}

################################################
## processed data stream read permissions
################################################

module "iam_kinesis_processed_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_processed.arn,
  ]
}

module "iam_kinesis_processed_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_kinesis_processed_read"
  policy = module.iam_kinesis_processed_read.kinesis_read_policy
  roles = [
    module.lambda_dynamodb_sink.role,
    module.lambda_processed_s3_sink.role,
  ]
}

################################################
## processed data stream write permissions
################################################

module "iam_kinesis_processed_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_processed.arn,
  ]
}

module "iam_kinesis_processed_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_kinesis_processed_write"
  policy = module.iam_kinesis_processed_write.kinesis_write_policy
  roles = [
    module.lambda_processor.role,
  ]
}
