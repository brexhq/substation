################################################
# DynamoDB table
# metadata is written from the processed Kinesis stream and read by the processor Lambda
################################################

module "dynamodb_example_sink" {
  source     = "./modules/dynamodb"
  kms_arn    = module.kms_substation.arn
  table_name = "substation_example"
  hash_key   = "PK"
  attributes = [
    {
      name = "PK"
      type = "S"
    }
  ]
}

################################################
# Lambda
# reads from processed Kinesis stream, writes to DynamoDB table
################################################

module "lambda_example_dynamodb_sink" {
  source        = "./modules/lambda"
  function_name = "substation_example_dynamodb_sink"
  description   = "Substation Lambda that is triggered from the processed Kinesis stream and writes data to DynamoDB"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest2"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_dynamodb_sink"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "example"
  }
}

resource "aws_lambda_event_source_mapping" "lambda_esm_example_dynamodb_sink" {
  event_source_arn                   = module.kinesis_example_processed.arn
  function_name                      = module.lambda_example_dynamodb_sink.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

################################################
## permissions
################################################

module "iam_lambda_example_dynamodb_sink_write" {
  source = "./modules/iam"
  resources = [
    module.dynamodb_example_sink.arn,
  ]
}

module "iam_lambda_example_dynamodb_sink_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "${module.lambda_example_dynamodb_sink.name}_write"
  policy = module.iam_lambda_example_dynamodb_sink_write.dynamodb_write_policy
  roles = [
    module.lambda_example_dynamodb_sink.role
  ]
}
