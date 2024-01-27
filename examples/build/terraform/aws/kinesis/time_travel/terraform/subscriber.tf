module "lambda_subscriber" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "subscriber"
    description = "Substation node that reads from Kinesis with a delay to support enrichment"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/subscriber"
      "SUBSTATION_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_event_source_mapping" "lambda_subscriber" {
  event_source_arn                   = module.kinesis.arn
  function_name                      = module.lambda_subscriber.arn
  maximum_batching_window_in_seconds = 15
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
