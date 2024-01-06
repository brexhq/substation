output "role" {
  value = aws_iam_role.destination
}

output "arn" {
  value = aws_cloudwatch_log_destination.destination.arn
}
