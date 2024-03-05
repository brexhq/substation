module "lambda_enrichment" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "enrichment"
    description = "Substation node that enriches data from Kinesis and writes it to DynamoDB"
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/enrichment"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_event_source_mapping" "lambda_enrichment" {
  event_source_arn                   = module.kinesis.arn
  function_name                      = module.lambda_enrichment.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
