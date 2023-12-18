resource "aws_cloudwatch_log_subscription_filter" "subscription_filter" {
  for_each       = toset(var.config.log_groups)
  log_group_name = each.key

  name            = var.config.name
  destination_arn = var.config.destination_arn
  # By default there is no filter pattern, so all logs are sent to the destination.
  filter_pattern = var.config.filter_pattern

  # If the destination is a Kinesis stream, then randomly distribute the logs to avoid hot shards.
  distribution = "Random"
}
