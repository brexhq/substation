module "lambda_enrichment" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "enrichment"
    description = "Substation enrichment node that writes threat signals to DynamoDB."
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/enrichment"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "lambda_enrichment" {
  event_source_arn                   = module.kinesis.arn
  function_name                      = module.lambda_enrichment.arn
  maximum_batching_window_in_seconds = 10
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
