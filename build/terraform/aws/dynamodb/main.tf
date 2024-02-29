resource "random_uuid" "id" {}

locals {
  read_capacity = var.config.read_capacity != null ? var.config.read_capacity : tomap({
    min    = 5
    max    = 1000
    target = 70
  })

  write_capacity = var.config.write_capacity != null ? var.config.write_capacity : tomap({
    min    = 5
    max    = 1000
    target = 70
  })
}

resource "aws_dynamodb_table" "table" {
  name           = var.config.name
  billing_mode   = "PROVISIONED"
  read_capacity  = local.read_capacity.min
  write_capacity = local.write_capacity.min
  hash_key       = var.config.hash_key
  range_key      = var.config.range_key

  # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html.
  ttl {
    attribute_name = var.config.ttl != null ? var.config.ttl : ""
    enabled        = var.config.ttl != null ? true : false
  }
  point_in_time_recovery {
    enabled = true
  }
  lifecycle {
    ignore_changes = [read_capacity, write_capacity]
  }

  server_side_encryption {
    enabled     = var.kms != null ? true : false
    kms_key_arn = var.kms != null ? var.kms.arn : null
  }

  # Streams are only charged for read operations and reads from AWS Lambda are free:
  # https://aws.amazon.com/dynamodb/pricing/.
  stream_enabled   = var.config.stream_view_type != null ? true : false
  stream_view_type = var.config.stream_view_type

  dynamic "attribute" {
    for_each = var.config.attributes

    content {
      name = attribute.value.name
      type = attribute.value.type
    }
  }

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-dynamodb-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} DynamoDB table."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      # Read actions
      "dynamodb:GetItem",
      "dynamodb:Query",
      # Write actions
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
    ]

    resources = [
      aws_dynamodb_table.table.arn,
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      # Read actions
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:ListStreams",
    ]

    resources = [
      aws_dynamodb_table.table.stream_arn,
    ]
  }

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
}

# read autoscaling
resource "aws_appautoscaling_target" "read_target" {
  max_capacity       = local.read_capacity.max
  min_capacity       = local.read_capacity.min
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

    target_value = local.read_capacity.target
  }
}

# write autoscaling
resource "aws_appautoscaling_target" "write_target" {
  max_capacity       = local.write_capacity.max
  min_capacity       = local.write_capacity.min
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

    target_value = local.write_capacity.target
  }
}
