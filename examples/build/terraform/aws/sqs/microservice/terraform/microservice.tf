module "microservice" {
  source = "../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "microservice"
    description = "Substation node that acts as an asynchronous microservice"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 10
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/microservice"
      "SUBSTATION_HANDLER" : "AWS_SQS"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_event_source_mapping" "microservice" {
  event_source_arn                   = module.sqs.arn
  function_name                      = module.microservice.arn
  maximum_batching_window_in_seconds = 10
  batch_size                         = 100
}
