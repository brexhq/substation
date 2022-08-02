################################################
# raw data stream
################################################

module "kinesis_example_raw" {
  source            = "/workspaces/substation/build/terraform/aws/kinesis"
  kms_key_id        = module.kms_substation.arn
  stream_name       = "substation_example_raw"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn

  tags = {
    "Owner" = "example"
  }
}

################################################
## raw data stream read permissions
################################################

module "iam_example_kinesis_raw_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_example_raw.arn,
  ]
}

module "iam_example_kinesis_raw_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kinesis_raw_read"
  policy = module.iam_example_kinesis_raw_read.kinesis_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_example_processor.role,
    module.lambda_example_raw_s3_sink.role,
  ]
}

################################################
## raw data stream write permissions
## all source Lambda must have this policy
################################################

module "iam_example_kinesis_raw_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_example_raw.arn,
  ]
}

module "iam_example_kinesis_raw_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kinesis_raw_write"
  policy = module.iam_example_kinesis_raw_write.kinesis_write_policy
  roles = [
    module.lambda_example_gateway_source.role,
    module.lambda_example_s3_source.role,
    module.example_gateway_kinesis.role,
  ]
}

################################################
# processed data stream
################################################

module "kinesis_example_processed" {
  source            = "/workspaces/substation/build/terraform/aws/kinesis"
  kms_key_id        = module.kms_substation.arn
  stream_name       = "substation_example_processed"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
}

################################################
## processed data stream read permissions
## all sink Lambda must have this policy
################################################

module "iam_example_kinesis_processed_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_example_processed.arn,
  ]
}

module "iam_example_kinesis_processed_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kinesis_processed_read"
  policy = module.iam_example_kinesis_processed_read.kinesis_read_policy
  roles = [
    module.lambda_example_dynamodb_sink.role,
    module.lambda_example_processed_s3_sink.role,
  ]
}

################################################
## processed data stream write permissions
## by default, only the processor Lambda should have this policy
################################################

module "iam_example_kinesis_processed_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    module.kinesis_example_processed.arn,
  ]
}

module "iam_example_kinesis_processed_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_kinesis_processed_write"
  policy = module.iam_example_kinesis_processed_write.kinesis_write_policy
  roles = [
    module.lambda_example_processor.role,
  ]
}
