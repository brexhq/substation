data "aws_caller_identity" "current" {}

resource "random_uuid" "id" {}

resource "aws_cloudwatch_event_rule" "rule" {
  name                = var.config.name
  description         = var.config.description
  event_bus_name      = var.config.event_bus_arn != null ? var.config.event_bus_arn : "default"
  schedule_expression = var.config.schedule != null ? var.config.schedule : null
  event_pattern       = var.config.event_pattern != null ? var.config.event_pattern : null
  tags                = var.tags
}

resource "aws_cloudwatch_event_target" "target" {
  rule      = aws_cloudwatch_event_rule.rule.name
  target_id = var.config.name
  arn       = var.config.function.arn
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = var.config.function.name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.rule.arn
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-eventbridge-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} EventBridge rule."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  # Always allow access to the default event bus for the account.
  statement {
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]

    resources = [
      "arn:aws:events:*:${data.aws_caller_identity.current.account_id}:event-bus/default",
    ]
  }

  dynamic "statement" {
    for_each = var.config.event_bus_arn != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "events:PutEvents",
      ]

      resources = [
        var.config.event_bus_arn,
      ]
    }
  }
}
