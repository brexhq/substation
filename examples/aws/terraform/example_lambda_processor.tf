################################################
# Lambda
# reads from raw Kinesis stream, writes to processed Kinesis stream
################################################

module "lambda_example_processor" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_example_processor"
  description   = "Substation Lambda that is triggered from the raw Kinesis stream and writes data to the processed Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_processor"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "example"
  }
}

resource "aws_lambda_event_source_mapping" "lambda_esm_example_processor" {
  event_source_arn                   = module.kinesis_example_raw.arn
  function_name                      = module.lambda_example_processor.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

################################################
## permissions
################################################

# allows processor Lambda to read from DynamoDB tables for enrichment
module "iam_lambda_example_processor_dynamodb_read" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "iam_lambda_example_processor_dynamodb_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "${module.lambda_example_processor.name}_dynamodb_read"
  policy = module.iam_lambda_example_processor_dynamodb_read.dynamodb_read_policy
  roles = [
    module.lambda_example_processor.role
  ]
}

# allows processor Lambda to execute Lambda for enrichment
module "iam_lambda_example_processor_lambda_execute" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "iam_lambda_example_processor_lambda_execute_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "${module.lambda_example_processor.name}_lambda_execute"
  policy = module.iam_lambda_example_processor_lambda_execute.lambda_execute_policy
  roles = [
    module.lambda_example_processor.role
  ]
}
