################################################
# API Gateway
# sends data to Lambda
################################################

module "gateway_example_lambda_source" {
  source       = "/workspaces/substation/build/terraform/aws/api_gateway/lambda"
  name         = "substation_lambda_example"
  function_arn = module.lambda_example_gateway_source.arn
}

################################################
# Lambda
# reads from API Gateway, writes to raw Kinesis stream
################################################

module "lambda_example_gateway_source" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_example_gateway_source"
  description   = "Substation Lambda that is triggered from an API Gateway and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_gateway_source"
    "SUBSTATION_HANDLER" : "GATEWAY"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }
}
