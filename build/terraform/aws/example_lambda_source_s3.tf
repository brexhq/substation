################################################
# S3 bucket
# stores objects that are read and ingested
################################################

module "s3_example_source" {
  source  = "./modules/s3"
  kms_arn = module.kms_substation.arn
  bucket  = "substation-example-source"
}

################################################
# Lambda
# reads from S3 bucket, writes to raw Kinesis stream
################################################

module "lambda_example_s3_source" {
  source        = "./modules/lambda"
  function_name = "substation_example_s3_source"
  description   = "Substation Lambda that is triggered from S3 and writes data to the raw Kinesis stream"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest2"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_s3_source"
    "SUBSTATION_HANDLER" : "S3"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}


resource "aws_s3_bucket_notification" "lambda_notification_example_s3_source" {
  bucket = module.s3_example_source.id

  lambda_function {
    lambda_function_arn = module.lambda_example_s3_source.arn
    events              = ["s3:ObjectCreated:*"]
    # enable prefix and suffix filtering based on the source service that is writing objects to the bucket
    # filter_prefix       = var.filter_prefix
    # filter_suffix       = var.filter_suffix
  }

  depends_on = [
    aws_lambda_permission.lambda_example_s3_source,
  ]
}

################################################
## permissions
################################################

resource "aws_lambda_permission" "lambda_example_s3_source" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_example_s3_source.name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3_example_source.arn
}

module "iam_lambda_example_s3_source_read" {
  source = "./modules/iam"
  resources = [
    "${module.s3_example_source.arn}/*",
  ]
}

module "iam_lambda_example_s3_source_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "${module.lambda_example_s3_source.name}_s3_read"
  policy = module.iam_lambda_example_s3_source_read.s3_read_policy
  roles = [
    module.lambda_example_s3_source.role,
  ]
}
