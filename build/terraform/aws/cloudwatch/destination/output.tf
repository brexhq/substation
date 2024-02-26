output "role" {
  value       = aws_iam_role.destination
  description = "The IAM role used by the CloudWatch destination."
}

output "arn" {
  value       = aws_cloudwatch_log_destination.destination.arn
  description = "The ARN of the CloudWatch destination."
}
