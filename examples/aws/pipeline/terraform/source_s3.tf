module "lambda_source_s3" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_source_s3"
    description = "Writes to Kinesis"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_source_s3"
      "SUBSTATION_HANDLER" : "AWS_S3"
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

resource "aws_lambda_permission" "lambda_source_s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_source_s3.name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3.arn
}

resource "aws_s3_bucket_notification" "lambda_source_s3" {
  bucket = module.s3.id

  lambda_function {
    lambda_function_arn = module.lambda_source_s3.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix = "source/"
  }

  depends_on = [
    aws_lambda_permission.lambda_source_s3,
  ]
}
