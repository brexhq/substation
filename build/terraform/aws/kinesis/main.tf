resource "aws_kinesis_stream" "stream" {
  name             = var.stream_name
  shard_count      = var.shard_count
  retention_period = var.retention_period
  encryption_type  = "KMS"
  kms_key_id       = var.kms_key_id
  lifecycle {
    ignore_changes = [shard_count]
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "metric_alarm_downscale" {
  alarm_name          = "${var.stream_name}_downscale"
  alarm_description   = var.stream_name
  actions_enabled     = true
  alarm_actions       = [var.autoscaling_topic]
  evaluation_periods  = 120
  datapoints_to_alarm = 114
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
        "StreamName" = var.stream_name
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
        "StreamName" = var.stream_name
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
    expression  = "e1/(1000*60*${var.shard_count})"
    label       = "IncomingRecordsPercent"
    return_data = false
  }

  metric_query {
    id          = "e4"
    expression  = "e2/(1048576*60*${var.shard_count})"
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
  alarm_name          = "${var.stream_name}_upscale"
  alarm_description   = var.stream_name
  actions_enabled     = true
  alarm_actions       = [var.autoscaling_topic]
  evaluation_periods  = 10
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
        "StreamName" = var.stream_name
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
        "StreamName" = var.stream_name
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
    expression  = "e1/(1000*60*${var.shard_count})"
    label       = "IncomingRecordsPercent"
    return_data = false
  }

  metric_query {
    id          = "e4"
    expression  = "e2/(1048576*60*${var.shard_count})"
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
