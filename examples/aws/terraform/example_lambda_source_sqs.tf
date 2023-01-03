################################################
# SQS queue
# sends data to Lambda
################################################

module "sqs_source" {
  source     = "/workspaces/substation/build/terraform/aws/sqs"
  kms_key_id = module.kms_substation.key_id
  name       = "substation_sqs_example"
  # timeout must match timeout on Lambda
  visibility_timeout_seconds = 300
}

################################################
# Lambda
# reads from SQS queue, writes to raw Kinesis stream
################################################

module "lambda_sqs_source" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_sqs_source"
  description   = "Substation Lambda that is triggered from SQS and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]
  timeout       = 300

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_sqs_source"
    "SUBSTATION_HANDLER" : "AWS_SQS"
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

resource "aws_lambda_event_source_mapping" "lambda_esm_sqs_source" {
  event_source_arn                   = module.sqs_source.arn
  function_name                      = module.lambda_sqs_source.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
}

################################################
## permissions
################################################

module "iam_lambda_sqs_source_read" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    "${module.sqs_source.arn}",
  ]
}

module "iam_lambda_sqs_source_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "${module.lambda_sqs_source.name}_read"
  policy = module.iam_lambda_sqs_source_read.sqs_read_policy
  roles = [
    module.lambda_sqs_source.role
  ]
}
