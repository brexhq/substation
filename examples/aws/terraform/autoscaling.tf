# Used for deploying and maintaining the Kinesis Data Streams autoscaling application; does not need to be used if deployments don't include Kinesis Data Streams.

resource "aws_sns_topic" "autoscaling_topic" {
  name              = "substation_autoscaling"
  kms_master_key_id = module.kms_substation.key_id

  tags = {
    owner = "example"
  }
}

# first runs of this Terraform will fail due to an empty ECR image
module "lambda_autoscaling" {
  source        = "/workspaces/substation/build/terraform/aws/lambda"
  function_name = "substation_autoscaling"
  description   = "Autoscales Kinesis streams based on data volume and size"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_autoscaling.repository_url}:latest"
  architectures = ["arm64"]

  env = {
    AWS_APPCONFIG_EXTENSION_PREFETCH_LIST : "/applications/substation/environments/prod/configurations/substation_autoscaling"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.repository_url,
  ]
}

resource "aws_sns_topic_subscription" "autoscaling_subscription" {
  topic_arn = aws_sns_topic.autoscaling_topic.arn
  protocol  = "lambda"
  endpoint  = module.lambda_autoscaling.arn

  depends_on = [
    module.lambda_autoscaling.name
  ]
}

resource "aws_lambda_permission" "autoscaling_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_autoscaling.name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.autoscaling_topic.arn

  depends_on = [
    module.lambda_autoscaling.name
  ]
}

# required for updating shard counts on Kinesis streams
# resources can be isolated, but defaults to all streams
module "autoscaling_kinesis_modify" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "autoscaling_kinesis_modify_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_autoscaling_kinesis_modify_attachment"
  policy = module.autoscaling_kinesis_modify.kinesis_modify_policy
  roles = [
    module.lambda_autoscaling.role,
  ]
}

# required for reading active shard counts for Kinesis streams
# resources can be isolated, but defaults to all streams
module "autoscaling_kinesis_read" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "autoscaling_kinesis_read_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_autoscaling_kinesis_read_attachment"
  policy = module.autoscaling_kinesis_read.kinesis_read_policy
  roles = [
    module.lambda_autoscaling.role,
  ]
}

# required for resetting CloudWatch alarm states
# resources can be isolated, but defaults to all streams
module "autoscaling_cloudwatch_modify" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "autoscaling_cloudwatch_modify_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_autoscaling_cloudwatch_modify_attachment"
  policy = module.autoscaling_cloudwatch_modify.cloudwatch_modify_policy
  roles = [
    module.lambda_autoscaling.role,
  ]
}

# required for updating CloudWatch alarm variables
# resources can be isolated, but defaults to all streams
module "autoscaling_cloudwatch_write" {
  source    = "/workspaces/substation/build/terraform/aws/iam"
  resources = ["*"]
}

module "autoscaling_cloudwatch_write_attachment" {
  source = "/workspaces/substation/build/terraform/aws/iam_attachment"
  id     = "substation_autoscaling_cloudwatch_write_attachment"
  policy = module.autoscaling_cloudwatch_write.cloudwatch_write_policy
  roles = [
    module.lambda_autoscaling.role,
  ]
}
