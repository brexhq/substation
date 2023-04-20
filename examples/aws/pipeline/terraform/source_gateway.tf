################################################
# API Gateway
# sends data to raw Kinesis stream
################################################

module "gateway_kinesis_source" {
  source = "../../../../build/terraform/aws/api_gateway/kinesis"
  name   = "substation_kinesis_example"
  stream = "substation_raw"
}

################################################
# API Gateway
# sends data to Lambda
################################################

module "gateway_lambda_source" {
  source       = "../../../../build/terraform/aws/api_gateway/lambda"
  name         = "substation_lambda_example"
  function_arn = module.lambda_gateway_source.arn
}

################################################
# Lambda
# reads from API Gateway, writes to raw Kinesis stream
################################################

module "lambda_gateway_source" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_gateway_source"
  description   = "Substation Lambda that is triggered from an API Gateway and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_gateway_source"
    "SUBSTATION_HANDLER" : "AWS_API_GATEWAY"
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
