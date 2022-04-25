# example of full-featured data pipeline, see README for more detail
# core infrastructure

# objects put into this S3 bucket are ingested into the data pipeline
# as a best practice, this data should not be modified before it is put into a raw Kinesis stream

data "aws_region" "current" {}


module "example_s3_source" {
  source  = "./modules/s3"
  kms_arn = module.substation_kms.arn
  bucket  = "substation-example-source"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = module.example_s3_source.id

  lambda_function {
    lambda_function_arn = module.example_s3_source_lambda.arn
    events              = ["s3:ObjectCreated:*"]
    # enable prefix and suffix filtering based on the source service that is writing objects to the bucket
    # filter_prefix       = var.filter_prefix
    # filter_suffix       = var.filter_suffix
  }

  depends_on = [aws_lambda_permission.allow_buckets]
}

# WORM bucket is used to store ummodified, raw data ingested into the data pipeline
# as a best practice, use this when you must store data due to compliance requirements
# in this example we've commented out the worm module and replaced it with the standard s3 module, otherwise you won't be able to delete the WORM bucket :)
module "example_s3_worm" {
  # source  = "./modules/s3/worm"
  source  = "./modules/s3"
  kms_arn = module.substation_kms.arn
  bucket  = "substation-example-worm"
}

# objects put into this S3 bucket are ready for use by analytic applications
module "example_s3_sink" {
  source  = "./modules/s3"
  kms_arn = module.substation_kms.arn
  bucket  = "substation-example-sink"
}

# data sent to this API Gateway is written directly to the raw Kinesis stream
module "example_gateway_kinesis" {
  source = "./modules/api_gateway/kinesis"
  name   = "substation_kinesis_example"
  stream = "substation_raw_example"
}

# unmodified, raw data is stored in this Kinesis stream
# as a best practice, data transformation Lambda should read from this stream
module "example_kinesis_raw" {
  source            = "./modules/kinesis"
  kms_key_id        = module.substation_kms.key_id
  stream_name       = "substation_raw_example"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
}

# modified, processed data is stored in this Kinesis stream
# as a best practice, data transformation Lambda should write to this stream
module "example_kinesis_processed" {
  source            = "./modules/kinesis"
  kms_key_id        = module.substation_kms.key_id
  stream_name       = "substation_processed_example"
  autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
}

/*
data put into this table can be read by Substation Lambda using the DynamoDB processor or by other AWS applications 
best practices:
- use one table per data pipeline (e.g., cloudtrail) or one table per use case (e.g., passive DNS database)
- name the hash/partition key "pk" and sort key "sk"
  - this provides flexibility to reuse a single table for multiple use cases
*/
module "example_dynamodb" {
  source     = "./modules/dynamodb"
  kms_arn    = module.substation_kms.arn
  table_name = "substation_example"
  hash_key   = "pk"
  attributes = [
    {
      name = "pk"
      type = "S"
    }
  ]
}

# lambda

module "example_s3_source_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_s3_source"
  description = "Example Substation Lambda that reads data from an S3 bucket and writes it to the raw Kinesis stream"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    # any Lambda that sink to Kinesis or DynamoDB should set AWS_MAX_ATTEMPTS to a moderately high value -- for example, between 10 and 20 -- so that the Lambda doesn't fail during autoscaling
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_s3_source"
    "SUBSTATION_HANDLER" : "S3"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

resource "aws_lambda_permission" "allow_buckets" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.example_s3_source_lambda.arn
  principal     = "s3.amazonaws.com"
  source_arn    = module.example_s3_source.arn
}

module "example_s3_worm_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_s3_worm"
  description = "Example Substation Lambda that reads data from the raw Kinesis stream and writes it to the WORM S3 bucket"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_s3_worm"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

module "example_s3_sink_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_s3_sink"
  description = "Example Substation Lambda that reads data from the processed Kinesis stream and writes it to an S3 bucket"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn
  
  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_s3_sink"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

module "example_gateway_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_gateway"
  description = "Example Substation Lambda that reads data from API Gateway and writes it to the raw Kinesis stream"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_gateway"
    "SUBSTATION_HANDLER" : "GATEWAY"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

# data sent to this API Gateway is sent to a Lambda for processing
module "example_gateway" {
  source       = "./modules/api_gateway/lambda"
  name         = "substation_example_lambda"
  function_arn = module.example_gateway_lambda.arn
}

module "example_kinesis_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_kinesis"
  description = "Example Substation Lambda that reads data from the raw Kinesis stream and writes it to the processed Kinesis stream"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_kinesis"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

module "example_dynamodb_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_example_dynamodb"
  description = "Example Substation Lambda that reads data from the processed Kinesis data stream and writes it to DynamoDB"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_example_dynamodb"
    "SUBSTATION_HANDLER" : "KINESIS"
    "SUBSTATION_DEBUG" : 1
  }
  tags = {
    "Owner" = "substation_example"
  }
  image_uri = "${module.substation_ecr.repository_url}:latest"
}

# IAM policies and attachments

# every Lambda should have AppConfig read access
module "example_appconfig_read" {
  source    = "./modules/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

module "example_appconfig_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_example_appconfig_read"
  policy = module.example_appconfig_read.appconfig_read_policy
  roles = [
    module.example_s3_source_lambda.role,
    module.example_s3_worm_lambda.role,
    module.example_s3_sink_lambda.role,
    module.example_dynamodb_lambda.role,
    module.example_kinesis_lambda.role,
    module.example_gateway_lambda.role,
  ]
}

module "example_kms_read" {
  source    = "./modules/iam"
  resources = [module.substation_kms.arn, aws_kms_key.xray_key.arn]
}

# every Lambda should have KMS read access
module "example_kms_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kms_read"
  policy = module.example_kms_read.kms_read_policy
  roles = [
    module.example_s3_source_lambda.role,
    module.example_s3_worm_lambda.role,
    module.example_s3_sink_lambda.role,
    module.example_dynamodb_lambda.role,
    module.example_kinesis_lambda.role,
    module.example_gateway_lambda.role,
  ]
}

# every Lambda that interacts with an encrypted resource needs KMS write access
module "example_kms_write" {
  source    = "./modules/iam"
  resources = [module.substation_kms.arn, aws_kms_key.xray_key.arn]
}

module "example_kms_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kms_write"
  policy = module.example_kms_write.kms_write_policy
  roles = [
    module.example_s3_source_lambda.role,
    module.example_s3_worm_lambda.role,
    module.example_s3_sink_lambda.role,
    module.example_dynamodb_lambda.role,
    module.example_kinesis_lambda.role,
    module.example_gateway_lambda.role,
    module.example_gateway_kinesis.role,
  ]
}

module "example_s3_source_read" {
  source = "./modules/iam"
  resources = [
    "${module.example_s3_source.arn}/*",
  ]
}

module "example_s3_source_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_s3_source_read"
  policy = module.example_s3_source_read.s3_read_policy
  roles = [
    module.example_s3_source_lambda.role
  ]
}

module "example_s3_worm_write" {
  source = "./modules/iam"
  resources = [
    "${module.example_s3_worm.arn}/*"
  ]
}

module "example_s3_worm_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_s3_worm_write"
  policy = module.example_s3_worm_write.s3_write_policy
  roles = [
    module.example_s3_worm_lambda.role
  ]
}

module "example_s3_sink_write" {
  source = "./modules/iam"
  resources = [
    "${module.example_s3_sink.arn}/*"
  ]
}

module "example_s3_sink_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_s3_sink_write"
  policy = module.example_s3_sink_write.s3_write_policy
  roles = [
    module.example_s3_sink_lambda.role
  ]
}

# only the Kinesis Lambda needs access to invoke other Lambda (via Substation's Lambda processor)
module "example_lambda_execute" {
  source    = "./modules/iam"
  resources = ["*"]
}

module "example_lambda_execute_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_lambda_execute"
  policy = module.example_lambda_execute.lambda_execute_policy
  roles = [
    module.example_kinesis_lambda.role
  ]
}

module "example_kinesis_raw_read" {
  source    = "./modules/iam"
  resources = [module.example_kinesis_raw.arn]
}

module "example_kinesis_raw_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kinesis_raw_read"
  policy = module.example_kinesis_raw_read.kinesis_read_policy
  roles = [
    module.example_kinesis_lambda.role,
    module.example_s3_worm_lambda.role
  ]
}

module "example_kinesis_raw_write" {
  source    = "./modules/iam"
  resources = [module.example_kinesis_raw.arn]
}

module "example_kinesis_raw_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kinesis_raw_write"
  policy = module.example_kinesis_raw_write.kinesis_write_policy
  roles = [
    module.example_s3_source_lambda.role,
    module.example_gateway_lambda.role,
    module.example_gateway_kinesis.role,
  ]
}

module "example_kinesis_processed_read" {
  source    = "./modules/iam"
  resources = [module.example_kinesis_processed.arn]
}

module "example_kinesis_processed_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kinesis_processed_read"
  policy = module.example_kinesis_processed_read.kinesis_read_policy
  roles = [
    module.example_s3_sink_lambda.role,
    module.example_dynamodb_lambda.role,
  ]
}

module "example_kinesis_processed_write" {
  source    = "./modules/iam"
  resources = [module.example_kinesis_processed.arn]
}

module "example_kinesis_processed_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_kinesis_processed_write"
  policy = module.example_kinesis_processed_write.kinesis_write_policy
  roles = [
    module.example_kinesis_lambda.role,
  ]
}

module "example_dynamodb_write" {
  source    = "./modules/iam"
  resources = [module.example_dynamodb.arn]
}

module "example_dynamodb_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "example_dynamodb_write"
  policy = module.example_dynamodb_write.dynamodb_write_policy
  roles = [
    module.example_dynamodb_lambda.role
  ]
}

resource "aws_lambda_event_source_mapping" "example_s3_worm_lambda_mapping" {
  event_source_arn                   = module.example_kinesis_raw.arn
  function_name                      = module.example_s3_worm_lambda.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 10
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

resource "aws_lambda_event_source_mapping" "example_kinesis_lambda_mapping" {
  event_source_arn                   = module.example_kinesis_raw.arn
  function_name                      = module.example_kinesis_lambda.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 10
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

resource "aws_lambda_event_source_mapping" "example_s3_sink_lambda_mapping" {
  event_source_arn                   = module.example_kinesis_processed.arn
  function_name                      = module.example_s3_sink_lambda.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 10
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}

resource "aws_lambda_event_source_mapping" "example_dynamodb_lambda_mapping" {
  event_source_arn                   = module.example_kinesis_processed.arn
  function_name                      = module.example_dynamodb_lambda.arn
  maximum_batching_window_in_seconds = 5
  batch_size                         = 10
  parallelization_factor             = 1
  starting_position                  = "LATEST"
}
