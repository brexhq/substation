data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation resources.
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

  # Allows S3 to access encrypted SNS topic.
  statement {
    sid    = "S3"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
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
  source = "../../../../../../../build/terraform/aws/appconfig"

  config = {
    name        = "substation"
    environments = [{
      name = "example"
    }]
  }
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

# S3 bucket used to store all data.
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

module "sns" {
  source = "../../../../../../../build/terraform/aws/sns"
  kms    = module.kms

  config = {
    name = "substation"
  }
}

# Grants the S3 bucket permission to publish to the SNS topic.
resource "aws_sns_topic_policy" "s3_access" {
  arn    = module.sns.arn
  policy = data.aws_iam_policy_document.s3_access_policy.json
}

data "aws_iam_policy_document" "s3_access_policy" {
  statement {
    actions = [
      "sns:Publish",
    ]

    resources = [
      module.sns.arn,
    ]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"

      values = [
        module.s3.arn,
      ]
    }

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    effect = "Allow"
  }
}

resource "aws_s3_bucket_notification" "sns" {
  bucket = module.s3.id

  topic {
    topic_arn = module.sns.arn

    events = [
      "s3:ObjectCreated:*",
    ]
  }
}
