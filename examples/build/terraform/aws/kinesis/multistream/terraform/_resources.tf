data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation infrastructure
module "kms" {
  source = "../../../../../../../build/terraform/aws/kms"

  config = {
    name   = "alias/substation"
    policy = data.aws_iam_policy_document.kms.json
  }
}

# This policy is required to support encrypted SNS topics.
# More information: https://repost.aws/knowledge-center/cloudwatch-receive-sns-for-alarm-trigger
data "aws_iam_policy_document" "kms" {
  # Allows CloudWatch to access encrypted SNS topic.
  statement {
    sid    = "CloudWatch"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    principals {
      type        = "Service"
      identifiers = ["cloudwatch.amazonaws.com"]
    }

    resources = ["*"]
  }

  # Default key policy for KMS.
  # https://docs.aws.amazon.com/kms/latest/developerguide/determining-access-key-policy.html
  statement {
    sid    = "KMS"
    effect = "Allow"
    actions = [
      "kms:*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.caller.account_id}:root"]
    }

    resources = ["*"]
  }
}

# AppConfig application that is shared by all Substation applications.
resource "aws_appconfig_application" "substation" {
  name        = "substation"
  description = "Stores compiled configuration files for Substation"
}

resource "aws_appconfig_environment" "example" {
  name           = "example"
  description    = "Stores example Substation configuration files"
  application_id = aws_appconfig_application.substation.id
}

# AWS Lambda requires an instant deployment strategy.
resource "aws_appconfig_deployment_strategy" "instant" {
  name                           = "Instant"
  description                    = "This strategy deploys the configuration to all targets immediately with zero bake time."
  deployment_duration_in_minutes = 0
  final_bake_time_in_minutes     = 0
  growth_factor                  = 100
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
}

# Repository for the core Substation application.
module "ecr_substation" {
  source = "../../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "substation"
    force_delete = true
  }
}

# Repository for the autoscaling application.
module "ecr_autoscaling" {
  source = "../../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "autoscaler"
    force_delete = true
  }
}

# SNS topic for Kinesis Data Stream autoscaling alarms.
resource "aws_sns_topic" "autoscaling_topic" {
  name              = "autoscaler"
  kms_master_key_id = module.kms.id
}

# API Gateway that sends data to Kinesis.
module "gateway_to_kinesis" {
  source = "../../../../../../../build/terraform/aws/api_gateway/kinesis_data_stream"
  # Always required for the Kinisis Data Stream integration.
  kinesis_data_stream = module.kds_src

  config = {
    name = "gateway"
  }
}

# Kinesis Data Stream that stores data sent from pipeline sources.
module "kds_src" {
  source = "../../../../../../../build/terraform/aws/kinesis_data_stream"
  kms    = module.kms

  config = {
    name              = "substation_src"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Reads data to the stream.
    module.lambda_publisher.role.name,
    # Writes data to the stream.
    module.gateway_to_kinesis.role.name,
  ]
}

# Kinesis Data Stream that stores data sent from the pipeline processor.
module "kds_dst" {
  source = "../../../../../../../build/terraform/aws/kinesis_data_stream"
  kms    = module.kms

  config = {
    name              = "substation_dst"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Writes data to the stream.
    module.lambda_publisher.role.name,
    # Reads data from the stream.
    module.lambda_subscriber.role.name,
  ]
}
