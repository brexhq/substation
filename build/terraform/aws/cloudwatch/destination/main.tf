resource "random_uuid" "id" {}

data "aws_caller_identity" "current" {}

locals {
  # By default, the current account is included in the list of accounts.
  account_ids = concat(var.config.account_ids, [data.aws_caller_identity.current.account_id])
}

data "aws_iam_policy_document" "destination_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = [
        "logs.amazonaws.com"
      ]
    }

    condition {
      test     = "StringLike"
      variable = "aws:SourceArn"

      # Creates a list of wildcarded ARNs for each account.
      values = formatlist("arn:aws:logs:*:%s:*", local.account_ids)
    }
  }
}

data "aws_iam_policy_document" "destination" {
  dynamic "statement" {
    for_each = var.kms != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ]

      resources = [
        var.kms.arn,
      ]
    }
  }

  // If the destination is Kinesis Firehose, the role must have write access.
  dynamic "statement" {
    for_each = strcontains(var.config.destination_arn, "arn:aws:firehose:") ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "firehose:DescribeDeliveryStream",
        "firehose:PutRecord",
        "firehose:PutRecordBatch",
      ]

      resources = [
        var.config.destination_arn,
      ]
    }
  }

  // If the destination is Kinesis Data Stream, the role must have write access.
  dynamic "statement" {
    for_each = strcontains(var.config.destination_arn, "arn:aws:kinesis:") ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "kinesis:DescribeStream",
        "kinesis:DescribeStreamSummary",
        "kinesis:DescribeStreamConsumer",
        "kinesis:SubscribeToShard",
        "kinesis:RegisterStreamConsumer",
        "kinesis:PutRecord",
        "kinesis:PutRecords",
      ]

      resources = [
        var.config.destination_arn,
      ]
    }
  }
}

resource "aws_iam_role" "destination" {
  name               = "substation-cloudwatch-dest-${resource.random_uuid.id.id}"
  assume_role_policy = data.aws_iam_policy_document.destination_assume_role.json
  tags               = var.tags
}

resource "aws_iam_role_policy_attachment" "destination" {
  role       = aws_iam_role.destination.name
  policy_arn = aws_iam_policy.destination.arn
}

resource "aws_iam_policy" "destination" {
  name        = "substation-cloudwatch-dest-${resource.random_uuid.id.id}"
  description = "Policy for the ${var.config.name} CloudWatch destination."
  policy      = data.aws_iam_policy_document.destination.json
}

resource "aws_cloudwatch_log_destination" "destination" {
  name       = var.config.name
  role_arn   = aws_iam_role.destination.arn
  target_arn = var.config.destination_arn
}

data "aws_iam_policy_document" "destination_access" {
  statement {
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = local.account_ids
    }

    actions = [
      "logs:PutSubscriptionFilter",
    ]
    resources = [
      aws_cloudwatch_log_destination.destination.arn,
    ]
  }
}

resource "aws_cloudwatch_log_destination_policy" "destination" {
  destination_name = aws_cloudwatch_log_destination.destination.name
  access_policy    = data.aws_iam_policy_document.destination_access.json
}
