################################################
# Lambda
# reads from processed Kinesis stream, writes to S3
################################################

module "lambda_processed_s3_sink" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_processed_s3_sink"
  description   = "Substation Lambda that is triggered from the processed Kinesis stream and writes data to S3"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_processed_s3_sink"
    "SUBSTATION_HANDLER" : "AWS_KINESIS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.repository_url,
    module.network,
  ]
}

resource "aws_lambda_event_source_mapping" "lambda_esm_processed_s3_sink" {
  event_source_arn                   = module.kinesis_processed.arn
  function_name                      = module.lambda_processed_s3_sink.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
