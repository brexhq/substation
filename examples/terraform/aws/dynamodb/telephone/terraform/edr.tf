# Kinesis Data Stream that stores data sent from pipeline sources.
module "edr_kinesis" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"

  config = {
    name              = "substation_edr"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Consumes data from the stream.
    module.edr_enrichment.role.name,
    module.edr_transform.role.name,
  ]
}

module "edr_transform" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "edr_transform"
    description = "Substation node that transforms EDR data."
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/edr_transform"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "edr_transform" {
  event_source_arn = module.edr_kinesis.arn
  function_name    = module.edr_transform.arn
  # This is set to 30 seconds (compared to the other data sources
  # 5 seconds) to simulate the asynchronous arrival of data in a 
  # real-world scenario.
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  # In this example, we start from the beginning of the stream,
  # but in a prod environment, you may want to start from the end
  # of the stream to avoid processing old data ("LATEST").
  starting_position = "TRIM_HORIZON"
}

module "edr_enrichment" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "edr_enrichment"
    description = "Substation node that enriches EDR data."
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/edr_enrichment"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "edr_enrichment" {
  event_source_arn                   = module.edr_kinesis.arn
  function_name                      = module.edr_enrichment.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 100
  parallelization_factor             = 1
  # In this example, we start from the beginning of the stream,
  # but in a prod environment, you may want to start from the end
  # of the stream to avoid processing old data ("LATEST").
  starting_position = "TRIM_HORIZON"
}
