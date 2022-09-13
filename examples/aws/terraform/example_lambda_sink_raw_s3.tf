################################################
# S3 bucket
# data is written from the raw Kinesis stream to the bucket
################################################

module "s3_example_raw_sink" {
  source  = "/workspaces/substation/build/terraform/aws/s3"
  kms_arn = module.kms_substation.arn
  bucket  = "substation-example-raw"
}

################################################
# Lambda
# reads from raw Kinesis stream, writes to S3
################################################

module "lambda_example_raw_s3_sink" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_example_raw_s3_sink"
  description   = "Substation Lambda that is triggered from the raw Kinesis stream and writes data to S3"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_raw_s3_sink"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }
}

resource "aws_lambda_event_source_mapping" "lambda_esm_example_raw_s3_sink" {
  event_source_arn                   = module.kinesis_example_raw.arn
  function_name                      = module.lambda_example_raw_s3_sink.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 100
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

################################################
## permissions
################################################

module "iam_lambda_example_raw_s3_sink_write" {
  source = "/workspaces/substation/build/terraform/aws/iam"
  resources = [
    "${module.s3_example_raw_sink.arn}/*",
  ]
}

module "iam_lambda_example_raw_s3_sink_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "${module.lambda_example_raw_s3_sink.name}_write"
  policy = module.iam_lambda_example_raw_s3_sink_write.s3_write_policy
  roles = [
    module.lambda_example_raw_s3_sink.role
  ]
}
