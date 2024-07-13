resource "aws_cloudwatch_event_rule" "rule" {
  name                = var.config.name
  description         = var.config.description
  schedule_expression = var.config.schedule
  event_bus_name      = var.config.event.bus_name
  event_pattern       = var.config.event.pattern
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
