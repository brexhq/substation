module "lambda_source_sqs" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_source_sqs"
    description = "Writes to Kinesis"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]
    timeout       = 300

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_source_sqs"
      "SUBSTATION_HANDLER" : "AWS_SQS"
      "SUBSTATION_DEBUG" : true
    }
  }

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.url,
    module.vpc,
  ]
}

resource "aws_lambda_event_source_mapping" "lambda_source_sqs" {
  event_source_arn                   = module.sqs.arn
  function_name                      = module.lambda_source_sqs.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
}
