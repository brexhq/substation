resource "aws_dynamodb_table" "table" {
  name           = var.table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = var.read_capacity_min
  write_capacity = var.write_capacity_min
  hash_key       = var.hash_key
  range_key      = var.range_key

  # services can opt in to use TTL functionality at runtime
  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html
  ttl {
    attribute_name = "ttl"
    enabled        = true
  }
  point_in_time_recovery {
    enabled = true
  }
  server_side_encryption {
    enabled     = true
    kms_key_arn = var.kms_arn
  }
  lifecycle {
    ignore_changes = [read_capacity, write_capacity]
  }

  # Streams are only charged for read operations and reads from AWS Lambda are free
  # https://aws.amazon.com/dynamodb/pricing/
  stream_enabled = true
  stream_view_type = var.stream_view_type

  dynamic "attribute" {
    for_each = var.attributes

    content {
      name = attribute.value.name
      type = attribute.value.type
    }
  }

  tags = var.tags
}

# read autoscaling
resource "aws_appautoscaling_target" "read_target" {
  max_capacity       = var.read_capacity_max
  min_capacity       = var.read_capacity_min
  resource_id        = "table/${aws_dynamodb_table.table.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "read_policy" {
  name               = "DynamoDBReadCapacityUtilization:${aws_appautoscaling_target.read_target.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.read_target.resource_id
  scalable_dimension = aws_appautoscaling_target.read_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.read_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    target_value = var.read_capacity_target
  }
}

# write autoscaling
resource "aws_appautoscaling_target" "write_target" {
  max_capacity       = var.write_capacity_max
  min_capacity       = var.write_capacity_min
  resource_id        = "table/${aws_dynamodb_table.table.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "write_policy" {
  name               = "DynamoDBWriteCapacityUtilization:${aws_appautoscaling_target.write_target.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.write_target.resource_id
  scalable_dimension = aws_appautoscaling_target.write_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.write_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    target_value = var.write_capacity_target
  }
}
