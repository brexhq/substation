################################################
# appconfig permissions
# all Lambda must have this policy
################################################

module "iam_example_appconfig_read" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

module "iam_example_appconfig_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_example_appconfig_read"
  policy = module.iam_example_appconfig_read.appconfig_read_policy
  roles = [
    module.lambda_autoscaling.role,
    module.lambda_example_processor.role,
    module.lambda_example_dynamodb_sink.role,
    module.lambda_example_raw_s3_sink.role,
    module.lambda_example_processed_s3_sink.role,
    module.lambda_example_gateway_source.role,
    module.lambda_example_s3_source.role,
    module.lambda_example_sqs_source.role,
  ]
}
