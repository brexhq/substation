module "lambda_transform" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "transform"
    description = "Substation transform node that enriches events with threat information."
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/transform"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "lambda_transform" {
  event_source_arn                   = module.kinesis.arn
  function_name                      = module.lambda_transform.arn
  maximum_batching_window_in_seconds = 10
  batch_size                         = 100
  parallelization_factor             = 1
  # In this example, we start from the beginning of the stream,
  # but in a prod environment, you may want to start from the end
  # of the stream to avoid processing old data ("LATEST").
  starting_position                  = "TRIM_HORIZON"
}
