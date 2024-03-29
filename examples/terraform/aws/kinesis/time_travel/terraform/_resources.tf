data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation infrastructure
module "kms" {
  source = "../../../../../../build/terraform/aws/kms"

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

module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

# Repository for the core Substation application.
module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "substation"
    force_delete = true
  }
}

# Repository for the autoscaling application.
module "ecr_autoscale" {
  source = "../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "autoscale"
    force_delete = true
  }
}

# SNS topic for Kinesis Data Stream autoscaling alarms.
resource "aws_sns_topic" "autoscaling_topic" {
  name              = "autoscale"
  kms_master_key_id = module.kms.id
}

# API Gateway that sends data to Kinesis.
module "gateway" {
  source = "../../../../../../build/terraform/aws/api_gateway/kinesis_data_stream"
  # Always required for the Kinisis Data Stream integration.
  kinesis_data_stream = module.kinesis

  config = {
    name = "gateway"
  }
}

# Kinesis Data Stream that stores data sent from pipeline sources.
module "kinesis" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"
  kms    = module.kms

  config = {
    name              = "substation"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Consumes data from the stream.
    module.lambda_enrichment.role.name,
    module.lambda_subscriber.role.name,
    # Publishes data to the stream.
    module.gateway.role.name,
  ]
}

module "dynamodb" {
  source = "../../../../../../build/terraform/aws/dynamodb"
  kms    = module.kms

  config = {
    name      = "substation"
    hash_key  = "PK"
    range_key = "SK"
    ttl       = "TTL"

    attributes = [
      {
        name = "PK"
        type = "S"
      },
      {
        name = "SK"
        type = "S"
      },
    ]
  }

  access = [
    module.lambda_enrichment.role.name,
    module.lambda_subscriber.role.name,
  ]
}
