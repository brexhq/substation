################################################
# SNS topic
# sends data to Lambda
################################################

module "sns_example_source" {
  source     = "/workspaces/substation/build/terraform/aws/sns"
  kms_key_id = module.kms_substation.key_id
  name       = "substation_sns_example"
}

################################################
# Lambda
# reads from SNS topic, writes to raw Kinesis stream
################################################

module "lambda_example_sns_source" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_example_sns_source"
  description   = "Substation Lambda that is triggered from SNS and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]
  timeout       = 300

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_sns_source"
    "SUBSTATION_HANDLER" : "SNS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}

resource "aws_sns_topic_subscription" "lambda_subscription_example_sns_source" {
  topic_arn = module.sns_example_source.arn
  protocol  = "lambda"
  endpoint  = module.lambda_example_sns_source.arn

  depends_on = [
    module.lambda_example_sns_source.name
  ]
}

################################################
## permissions
################################################

resource "aws_lambda_permission" "lambda_example_sns_source_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_example_sns_source.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns_example_source.arn

  depends_on = [
    module.lambda_example_sns_source.name
  ]
}
