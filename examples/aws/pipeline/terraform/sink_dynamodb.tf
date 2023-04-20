################################################
# DynamoDB table
# metadata is written from the processed Kinesis stream and read by the processor Lambda
################################################

module "dynamodb_sink" {
  source     = "../../../../build/terraform/aws/dynamodb"
  kms_arn    = module.kms_substation.arn
  table_name = "substation"
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

module "lambda_dynamodb_sink" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_dynamodb_sink"
  description   = "Substation Lambda that is triggered from the processed Kinesis stream and writes data to DynamoDB"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_dynamodb_sink"
    "SUBSTATION_HANDLER" : "AWS_KINESIS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  vpc_config = {
    subnet_ids = [
      module.network.private_subnet_id,
      module.network.public_subnet_id,
    ]
    security_group_ids = [module.network.public_egress_security_group_id]
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.repository_url,
    module.network,
  ]
}

resource "aws_lambda_event_source_mapping" "lambda_esm_dynamodb_sink" {
  event_source_arn                   = module.kinesis_processed.arn
  function_name                      = module.lambda_dynamodb_sink.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

################################################
## permissions
################################################

module "iam_lambda_dynamodb_sink_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.dynamodb_sink.arn,
  ]
}

module "iam_lambda_dynamodb_sink_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "${module.lambda_dynamodb_sink.name}_write"
  policy = module.iam_lambda_dynamodb_sink_write.dynamodb_write_policy
  roles = [
    module.lambda_dynamodb_sink.role
  ]
}
