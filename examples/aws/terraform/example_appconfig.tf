################################################
# appconfig permissions
# all Lambda must have this policy
################################################

module "iam_appconfig_read" {
  source    = "../../../build/terraform/aws/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

module "iam_appconfig_read_attachment" {
  source = "../../../build/terraform/aws/iam_attachment"
  id     = "substation_appconfig_read"
  policy = module.iam_appconfig_read.appconfig_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_processor.role,
    module.lambda_dynamodb_sink.role,
    module.lambda_raw_s3_sink.role,
    module.lambda_processed_s3_sink.role,
    module.lambda_gateway_source.role,
    module.lambda_s3_source.role,
    module.lambda_sns_source.role,
    module.lambda_sqs_source.role,
  ]
}
