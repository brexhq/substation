################################################
# Lambda
# reads from S3 bucket, writes to raw Kinesis stream
################################################

module "lambda_s3_source" {
  source        = "../../../build/terraform/aws/lambda"
  function_name = "substation_s3_source"
  description   = "Substation Lambda that is triggered from S3 and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_s3_source"
    "SUBSTATION_HANDLER" : "AWS_S3"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}


resource "aws_s3_bucket_notification" "lambda_notification_s3_source" {
  bucket = module.s3_substation.id

  lambda_function {
    lambda_function_arn = module.lambda_s3_source.arn
    events              = ["s3:ObjectCreated:*"]
    # enable prefix and suffix filtering based on the source service that is writing objects to the bucket
    filter_prefix       = "source/"
    # filter_suffix       = var.filter_suffix
  }

  depends_on = [
    aws_lambda_permission.lambda_s3_source,
  ]
}

################################################
## permissions
################################################

resource "aws_lambda_permission" "lambda_s3_source" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_s3_source.name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3_substation.arn
}
