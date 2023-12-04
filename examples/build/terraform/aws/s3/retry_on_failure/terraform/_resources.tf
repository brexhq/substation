data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation resources.
module "kms" {
  source = "../../../../../../../build/terraform/aws/kms"
  config = {
    name   = "alias/substation"
    policy = data.aws_iam_policy_document.kms.json
  }
}

data "aws_iam_policy_document" "kms" {
  # Allows CloudWatch to access encrypted resources.
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
module "ecr" {
  source = "../../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "substation"
    force_delete = true
  }
}

resource "random_uuid" "s3" {}

# Monolithic S3 bucket used to store all data.
module "s3" {
  source = "../../../../../../../build/terraform/aws/s3"
  kms    = module.kms

  config = {
    # Bucket name is randomized to avoid collisions.
    name = "${random_uuid.s3.result}-substation"
  }

  access = [
    module.lambda_node.role.name,
  ]
}

module "sqs_queue" {
  source = "../../../../../../../build/terraform/aws/sqs"
  kms    = module.kms

  config = {
    name = "substation_retry_queue"
    # Delay for 30 seconds to allow the pipeline to recover.
    delay = 30
  }

  access = [
    module.lambda_node.role.name,
    module.lambda_retrier.role.name,
  ]
}
