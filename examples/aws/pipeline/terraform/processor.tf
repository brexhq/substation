################################################
# Lambda
# reads from raw Kinesis stream, writes to processed Kinesis stream
################################################

module "lambda_processor" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_processor"
    description = "Processes data between raw and processed Kinesis streams"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]

    # Lambda runs within a custom VPC.
    vpc_config = {
      subnet_ids         = module.vpc.private_subnet_id
      security_group_ids = [module.vpc.default_security_group_id]
    }

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_processor"
      "SUBSTATION_HANDLER" : "AWS_KINESIS_DATA_STREAM"
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

resource "aws_lambda_event_source_mapping" "lambda_esm_processor" {
  event_source_arn                   = module.kinesis_raw.arn
  function_name                      = module.lambda_processor.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
