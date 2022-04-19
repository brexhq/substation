# Used for deploying and maintaining the Kinesis Data Streams autoscaling application; does not need to be used if deployments don't include Kinesis Data Streams.

resource "aws_sns_topic" "autoscaling_topic" {
  name              = "substation_autoscaling"
  kms_master_key_id = module.substation_kms.key_id
  tags = {
    "Owner" : "substation_example",
  }
}

# required for reading autoscaling configurations from AppConfig
module "autoscaling_appconfig_read" {
  source    = "./modules/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

# required for updating shard counts on Kinesis streams
# resources can be isolated, but defaults to all streams
module "autoscaling_kinesis_modify" {
  source    = "./modules/iam"
  resources = ["*"]
}

# required for reading active shard counts for Kinesis streams
# resources can be isolated, but defaults to all streams
module "autoscaling_kinesis_read" {
  source    = "./modules/iam"
  resources = ["*"]
}

# required for resetting CloudWatch alarm states
# resources can be isolated, but defaults to all streams
module "autoscaling_cloudwatch_modify" {
  source    = "./modules/iam"
  resources = ["*"]
}

# required for updating CloudWatch alarm variables
# resources can be isolated, but defaults to all streams
module "autoscaling_cloudwatch_write" {
  source    = "./modules/iam"
  resources = ["*"]
}

# first runs of this Terraform will fail due to an empty ECR image
module "autoscaling_lambda" {
  source               = "./modules/lambda"
  function_name        = "substation_autoscaling"
  description = "Autoscales Kinesis streams based on data volume and size"
  appconfig_id         = aws_appconfig_application.substation.id
  kms_arn = module.substation_kms.arn

  env = {
    AWS_APPCONFIG_EXTENSION_PREFETCH_LIST : "/applications/substation/environments/prod/configurations/autoscaling"
  }
  tags = {
    "Owner" : "substation_example",
  }
  image_uri = "${module.autoscaling_ecr.repository_url}:latest"
}

module "autoscaling_appconfig_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_autoscaling_appconfig_read_attachment"
  policy = module.autoscaling_appconfig_read.appconfig_read_policy
  roles  = [module.autoscaling_lambda.role]
}

module "autoscaling_kinesis_modify_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_autoscaling_kinesis_modify_attachment"
  policy = module.autoscaling_kinesis_modify.kinesis_modify_policy
  roles  = [module.autoscaling_lambda.role]
}

module "autoscaling_kinesis_read_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_autoscaling_kinesis_read_attachment"
  policy = module.autoscaling_kinesis_read.kinesis_read_policy
  roles  = [module.autoscaling_lambda.role]
}

module "autoscaling_cloudwatch_modify_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_autoscaling_cloudwatch_modify_attachment"
  policy = module.autoscaling_cloudwatch_modify.cloudwatch_modify_policy
  roles  = [module.autoscaling_lambda.role]
}

module "autoscaling_cloudwatch_write_attachment" {
  source = "./modules/iam_attachment"
  id     = "substation_autoscaling_cloudwatch_write_attachment"
  policy = module.autoscaling_cloudwatch_write.cloudwatch_write_policy
  roles  = [module.autoscaling_lambda.role]
}

resource "aws_lambda_permission" "autoscaling_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.autoscaling_lambda.name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.autoscaling_topic.arn
}

resource "aws_sns_topic_subscription" "autoscaling_subscription" {
  topic_arn = aws_sns_topic.autoscaling_topic.arn
  protocol  = "lambda"
  endpoint  = module.autoscaling_lambda.arn
}
