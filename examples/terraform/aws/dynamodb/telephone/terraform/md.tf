# Kinesis Data Stream that stores data sent from pipeline sources.
module "md_kinesis" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"

  config = {
    name              = "substation_md"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Consumes data from the stream.
    module.md_transform.role.name,
  ]
}

module "md_transform" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "md_transform"
    description = "Substation node that transforms managed device data."
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/md_transform"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }
}

resource "aws_lambda_event_source_mapping" "md_transform" {
  event_source_arn                   = module.md_kinesis.arn
  function_name                      = module.md_transform.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 100
  parallelization_factor             = 1
  # In this example, we start from the beginning of the stream,
  # but in a prod environment, you may want to start from the end
  # of the stream to avoid processing old data ("LATEST").
  starting_position = "TRIM_HORIZON"
}
