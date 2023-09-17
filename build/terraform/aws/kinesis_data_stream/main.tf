resource "aws_kinesis_stream" "stream" {
  name             = var.config.name
  shard_count      = var.config.shards
  retention_period = var.config.retention
  encryption_type  = "KMS"
  kms_key_id       = var.kms.id
  lifecycle {
    ignore_changes = [shard_count]
  }

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  for_each = toset(var.access)
  role = each.value
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = var.config.name
  description = "Policy for the ${var.config.name} Kinesis Data Stream"
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    sid = "KMS"

    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      var.kms.arn,
    ]
  }

  statement {
    sid = "CloudWatch"

    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
      "cloudwatch:PutMetricAlarm",
      "cloudwatch:SetAlarmState",
    ]

    resources = [
      aws_cloudwatch_metric_alarm.metric_alarm_downscale.arn,
      aws_cloudwatch_metric_alarm.metric_alarm_upscale.arn,
    ]
  }

  statement {
    sid = "Kinesis Data Stream"

    effect = "Allow"
    actions = [
      "kinesis:DescribeStream*",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:ListShards",
      "kinesis:ListStreams",
      "kinesis:PutRecord*",
      "kinesis:SubscribeToShard",
      "kinesis:SubscribeToShard",
      "kinesis:RegisterStreamConsumer",
      "kinesis:UpdateShardCount",
    ]

    resources = [
      aws_kinesis_stream.stream.arn,
    ]
  }
}


resource "aws_cloudwatch_metric_alarm" "metric_alarm_downscale" {
  alarm_name          = "${var.config.name}_downscale"
  alarm_description   = var.config.name
  actions_enabled     = true
  alarm_actions       = [var.config.autoscaling_topic]
  evaluation_periods  = 60
  datapoints_to_alarm = 57
  threshold           = 0.25
  comparison_operator = "LessThanOrEqualToThreshold"
  treat_missing_data  = "ignore"
  lifecycle {
    ignore_changes = [metric_query, datapoints_to_alarm]
  }

  metric_query {
    id = "m1"

    metric {
      namespace   = "AWS/Kinesis"
      metric_name = "IncomingRecords"
      dimensions = {
        "StreamName" = var.config.name
      }
      period = 60
      stat   = "Sum"
    }
    label       = "IncomingRecords"
    return_data = false
  }

  metric_query {
    id = "m2"

    metric {
      namespace   = "AWS/Kinesis"
      metric_name = "IncomingBytes"
      dimensions = {
        "StreamName" = var.config.name
      }
      period = 60
      stat   = "Sum"
    }
    label       = "IncomingBytes"
    return_data = false
  }

  metric_query {
    id          = "e1"
    expression  = "FILL(m1,0)"
    label       = "FillMissingIncomingRecords"
    return_data = false
  }

  metric_query {
    id          = "e2"
    expression  = "FILL(m2,0)"
    label       = "FillMissingIncomingBytes"
    return_data = false
  }

  metric_query {
    id          = "e3"
    expression  = "e1/(1000*60*${var.config.shards})"
    label       = "IncomingRecordsPercent"
    return_data = false
  }

  metric_query {
    id          = "e4"
    expression  = "e2/(1048576*60*${var.config.shards})"
    label       = "IncomingBytesPercent"
    return_data = false
  }

  metric_query {
    id          = "e5"
    expression  = "MAX([e3,e4])"
    label       = "IncomingMax"
    return_data = true
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "metric_alarm_upscale" {
  alarm_name          = "${var.config.name}_upscale"
  alarm_description   = var.config.name
  actions_enabled     = true
  alarm_actions       = [var.config.autoscaling_topic]
  evaluation_periods  = 5
  datapoints_to_alarm = 5
  threshold           = 0.75
  comparison_operator = "GreaterThanOrEqualToThreshold"
  treat_missing_data  = "ignore"
  lifecycle {
    ignore_changes = [metric_query, datapoints_to_alarm]
  }

  metric_query {
    id = "m1"

    metric {
      namespace   = "AWS/Kinesis"
      metric_name = "IncomingRecords"
      dimensions = {
        "StreamName" = var.config.name
      }
      period = 60
      stat   = "Sum"
    }
    label       = "IncomingRecords"
    return_data = false
  }

  metric_query {
    id = "m2"

    metric {
      namespace   = "AWS/Kinesis"
      metric_name = "IncomingBytes"
      dimensions = {
        "StreamName" = var.config.name
      }
      period = 60
      stat   = "Sum"
    }
    label       = "IncomingBytes"
    return_data = false
  }

  metric_query {
    id          = "e1"
    expression  = "FILL(m1,0)"
    label       = "FillMissingIncomingRecords"
    return_data = false
  }

  metric_query {
    id          = "e2"
    expression  = "FILL(m2,0)"
    label       = "FillMissingIncomingBytes"
    return_data = false
  }

  metric_query {
    id          = "e3"
    expression  = "e1/(1000*60*${var.config.shards})"
    label       = "IncomingRecordsPercent"
    return_data = false
  }

  metric_query {
    id          = "e4"
    expression  = "e2/(1048576*60*${var.config.shards})"
    label       = "IncomingBytesPercent"
    return_data = false
  }

  metric_query {
    id          = "e5"
    expression  = "MAX([e3,e4])"
    label       = "IncomingMax"
    return_data = true
  }

  tags = var.tags
}
