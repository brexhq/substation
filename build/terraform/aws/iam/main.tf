data "aws_iam_policy_document" "appconfig_read" {
  statement {
    effect = "Allow"
    actions = [
      "appconfig:GetConfiguration",
      "appconfig:GetLatestConfiguration",
      "appconfig:StartConfigurationSession",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "cloudwatch_modify" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:SetAlarmState",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "cloudwatch_write" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
      "cloudwatch:PutMetricAlarm",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kinesis_read" {
  statement {
    effect = "Allow"
    actions = [
      "kinesis:DescribeStream",
      "kinesis:DescribeStreamConsumer",
      "kinesis:DescribeStreamSummary",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:ListShards",
      "kinesis:ListStreams",
      "kinesis:SubscribeToShard",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kinesis_modify" {
  statement {
    effect = "Allow"
    actions = [
      "kinesis:UpdateShardCount",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kinesis_write" {
  statement {
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
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kinesis_firehose_write" {
  statement {
    effect = "Allow"
    actions = [
      "firehose:PutRecordBatch",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "dynamodb_read" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:Query",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "dynamodb_write" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
    ]
    resources = var.resources
  }
}

# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/iam-policy-read-stream-only.html
data "aws_iam_policy_document" "dynamodb_stream_read" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:DescribeStream",
      "dynamodb:ListStreams",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kms_read" {
  statement {
    effect = "Allow"
    actions = [
      "kms:Decrypt",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "kms_write" {
  statement {
    effect = "Allow"
    actions = [
      "kms:GenerateDataKey"
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "lambda_execute" {
  statement {
    effect = "Allow"
    actions = [
      "lambda:InvokeAsync",
      "lambda:InvokeFunction",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "s3_read" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "s3_write" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "secretsmanager_read" {
  statement {
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "sqs_read" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "sqs_write" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:GetQueueUrl",
      "sqs:SendMessage*",
    ]
    resources = var.resources
  }
}

data "aws_iam_policy_document" "sns_write" {
  statement {
    effect = "Allow"
    actions = [
      "sns:Publish",
    ]
    resources = var.resources
  }
}
