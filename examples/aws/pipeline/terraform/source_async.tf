################################################
# Lambda
# reads from async invocation, writes to raw Kinesis stream
################################################

module "lambda_async_source" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_source_async"
    description = "Writes to Kinesis"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_source_async"
      "SUBSTATION_HANDLER" : "AWS_LAMBDA_ASYNC"
      "SUBSTATION_DEBUG" : true
    }
  }

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.url,
  ]
}
