module "dynamodb" {
  source     = "../../../../build/terraform/aws/dynamodb"
  kms_arn    = module.kms.arn
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
# Substation node that publishes CDC events to SNS
################################################

module "publisher" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "publisher"
  description   = "Substation Lambda that publishes CDC events to SNS"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest3"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/publisher"
    "SUBSTATION_HANDLER" : "AWS_DYNAMODB"
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

resource "aws_lambda_event_source_mapping" "publisher" {
  event_source_arn  = module.dynamodb.stream_arn
  function_name     = module.publisher.arn
  starting_position = "LATEST"
}

################################################
# DynamoDB Stream read permissions
################################################

module "iam_dynamodb_stream_read" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.dynamodb.stream_arn,
  ]
}

module "iam_dynamodb_stream_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_dynamodb_stream_read"
  policy = module.iam_dynamodb_stream_read.dynamodb_stream_read_policy
  roles = [
    module.publisher.role,
  ]
}

################################################
# SNS write permissions
################################################

module "iam_sns_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.sns.arn,
  ]
}

module "iam_sns_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_sns_write"
  policy = module.iam_sns_write.sns_write_policy
  roles = [
    module.publisher.role,
  ]
}
